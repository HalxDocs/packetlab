package engine

import (
	"bytes"
	"testing"

	"github.com/HalxDocs/packetlab/internal/eventlog"
	"github.com/HalxDocs/packetlab/internal/network"
	"github.com/HalxDocs/packetlab/internal/sim"
	"github.com/HalxDocs/packetlab/internal/topology"
)

func encode(t *testing.T, e *Engine) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := eventlog.EncodeJSONL(&buf, e.Records()); err != nil {
		t.Fatalf("EncodeJSONL: %v", err)
	}
	return buf.Bytes()
}

func TestDeterministicReruns(t *testing.T) {
	topo := newTestTopology(t)
	opts := Options{Actions: []Action{pingAction()}}

	first := newTestEngine(t, topo, opts)
	runEngine(t, first)

	second := newTestEngine(t, topo, opts)
	runEngine(t, second)

	if !bytes.Equal(encode(t, first), encode(t, second)) {
		t.Fatal("two runs of the same scenario produced different event logs")
	}
}

func TestResetProducesIdenticalRun(t *testing.T) {
	e := newTestEngine(t, newTestTopology(t), Options{Actions: []Action{pingAction()}})
	runEngine(t, e)
	first := encode(t, e)

	e.Reset()
	records := e.Records()
	if len(records) != 1 || records[0].Kind != sim.RunStarted {
		t.Fatalf("after Reset records = %v, want only run.started", recordKinds(records))
	}
	if got := e.Result(); got.Processed != 0 || got.Delivered != 0 || got.Dropped != 0 || got.Completed {
		t.Fatalf("result after Reset = %+v", got)
	}

	runEngine(t, e)
	if !bytes.Equal(first, encode(t, e)) {
		t.Fatal("reset run did not reproduce the original event log")
	}
}

func TestEventLimit(t *testing.T) {
	e := newTestEngine(t, newTestTopology(t), Options{
		Actions:    []Action{pingAction()},
		EventLimit: 3,
	})
	if err := e.RunToCompletion(); err != nil {
		t.Fatalf("Run: %v", err)
	}
	result := e.Result()
	if result.Processed != 3 {
		t.Fatalf("Processed = %d, want 3", result.Processed)
	}
	if !result.LimitReached {
		t.Fatal("LimitReached = false, want true")
	}
	if result.Completed {
		t.Fatal("Completed = true, want false at the event limit")
	}
}

func TestStepProcessesOneEvent(t *testing.T) {
	e := newTestEngine(t, newTestTopology(t), Options{Actions: []Action{pingAction()}})

	ok, err := e.Step()
	if err != nil {
		t.Fatalf("Step: %v", err)
	}
	if !ok {
		t.Fatal("Step returned no event")
	}
	if got := e.Result().Processed; got != 1 {
		t.Fatalf("Processed = %d, want 1", got)
	}
	// run.started plus packet.created, route.selected, link.transmit.
	if got := len(e.Records()); got != 4 {
		t.Fatalf("records = %d, want 4 after one step", got)
	}
}

func TestSerializationDelay(t *testing.T) {
	link := &topology.Link{BandwidthBps: 8000}
	if got := serializationDelay(link, 1000); got != sim.Second {
		t.Fatalf("serializationDelay = %s, want 1s", got)
	}
	if got := serializationDelay(link, 500); got != 500*sim.Millisecond {
		t.Fatalf("serializationDelay = %s, want 500ms", got)
	}
	link.BandwidthBps = 0
	if got := serializationDelay(link, 1000); got != 0 {
		t.Fatalf("unlimited link delay = %s, want 0", got)
	}
}

func TestBandwidthDelaysArrival(t *testing.T) {
	topo := newTestTopology(t)
	link, _ := topo.Link("link-a")
	link.Delay = 0
	link.BandwidthBps = 264 // 33-byte packet takes exactly 1s

	e := newTestEngine(t, topo, Options{Actions: []Action{pingAction()}})
	runEngine(t, e)

	delivered, ok := findRecord(e.Records(), sim.LinkDelivered)
	if !ok {
		t.Fatal("no link.delivered record")
	}
	if delivered.At != sim.Second {
		t.Fatalf("first delivery at %s, want 1s", delivered.At)
	}
}

func TestActionValidation(t *testing.T) {
	topo := newTestTopology(t)
	base := pingAction()

	cases := map[string]Action{
		"unknown kind":        {At: 0, Kind: "send_udp", From: "host-a", To: base.To},
		"unknown source":      {At: 0, Kind: ActionSendPing, From: "host-z", To: base.To},
		"negative time":       {At: -1, Kind: ActionSendPing, From: "host-a", To: base.To},
		"invalid destination": {At: 0, Kind: ActionSendPing, From: "host-a"},
		"count too large":     {At: 0, Kind: ActionSendPing, From: "host-a", To: base.To, Count: 1001},
		"negative interval":   {At: 0, Kind: ActionSendPing, From: "host-a", To: base.To, Interval: -1},
	}
	for name, action := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := New(topo, Options{Actions: []Action{action}}); err == nil {
				t.Fatal("expected action validation error")
			}
		})
	}
}

func TestNewRejectsInvalidTopology(t *testing.T) {
	if _, err := New(topology.New(), Options{}); err == nil {
		t.Fatal("expected topology validation error")
	}
}

func TestNewRejectsNilTopology(t *testing.T) {
	if _, err := New(nil, Options{}); err == nil {
		t.Fatal("expected nil topology error")
	}
}

func TestPingCountSchedulesMultipleRequests(t *testing.T) {
	action := pingAction()
	action.Count = 3
	action.Interval = sim.Ms(10)

	e := newTestEngine(t, newTestTopology(t), Options{Actions: []Action{action}})
	runEngine(t, e)

	if got := e.Result().Delivered; got != 6 {
		t.Fatalf("Delivered = %d, want 6 (three request/reply pairs)", got)
	}
	var replies int
	for _, rec := range e.Records() {
		if rec.Kind == sim.PingReply {
			replies++
		}
	}
	if replies != 3 {
		t.Fatalf("ping.reply records = %d, want 3", replies)
	}
}

func TestSourceAddressComesFromEgressInterface(t *testing.T) {
	e := newTestEngine(t, newTestTopology(t), Options{Actions: []Action{pingAction()}})
	runEngine(t, e)

	created, ok := findRecord(e.Records(), sim.PacketCreated)
	if !ok {
		t.Fatal("no packet.created record")
	}
	if got := created.Details["src"]; got != "10.0.1.10" {
		t.Fatalf("src = %q, want 10.0.1.10", got)
	}
}

func TestPingToUnreachableNetwork(t *testing.T) {
	action := Action{
		At:   0,
		Kind: ActionSendPing,
		From: "host-a",
		To:   network.MustParseIPv4("192.168.50.1"),
	}
	e := newTestEngine(t, newTestTopology(t), Options{Actions: []Action{action}})
	runEngine(t, e)

	// host-a has a default route, so the packet reaches the router, which has
	// no route to 192.168.50.1 and drops it.
	drop, ok := findRecord(e.Records(), sim.PacketDropped)
	if !ok {
		t.Fatal("no packet.dropped record")
	}
	if drop.Reason != sim.ReasonNoRoute || drop.NodeID != "router-r1" {
		t.Fatalf("drop = %+v", drop)
	}
}
