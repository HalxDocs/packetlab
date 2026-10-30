package engine

import (
	"testing"

	"github.com/HalxDocs/packetlab/internal/eventlog"
	"github.com/HalxDocs/packetlab/internal/network"
	"github.com/HalxDocs/packetlab/internal/sim"
	"github.com/HalxDocs/packetlab/internal/topology"
)

// newTestTopology builds host-a — router-r1 — host-b with valid routes.
func newTestTopology(t *testing.T) *topology.Topology {
	t.Helper()
	topo := topology.New()

	nodes := []*topology.Node{
		{
			ID: "host-a", Kind: topology.KindHost,
			Routes: []network.Route{{
				Network:   network.MustParsePrefix("0.0.0.0/0"),
				NextHop:   network.MustParseIPv4("10.0.1.1"),
				Interface: "host-a:eth0",
			}},
		},
		{
			ID: "router-r1", Kind: topology.KindRouter,
			Routes: []network.Route{
				{Network: network.MustParsePrefix("10.0.1.0/24"), Interface: "router-r1:eth0"},
				{Network: network.MustParsePrefix("10.0.2.0/24"), Interface: "router-r1:eth1"},
			},
		},
		{
			ID: "host-b", Kind: topology.KindHost,
			Routes: []network.Route{{
				Network:   network.MustParsePrefix("0.0.0.0/0"),
				NextHop:   network.MustParseIPv4("10.0.2.1"),
				Interface: "host-b:eth0",
			}},
		},
	}
	for _, n := range nodes {
		if err := topo.AddNode(n); err != nil {
			t.Fatalf("AddNode %s: %v", n.ID, err)
		}
	}

	ifaces := []*topology.Interface{
		{ID: "host-a:eth0", Node: "host-a", Address: network.MustParsePrefix("10.0.1.10/24")},
		{ID: "router-r1:eth0", Node: "router-r1", Address: network.MustParsePrefix("10.0.1.1/24")},
		{ID: "router-r1:eth1", Node: "router-r1", Address: network.MustParsePrefix("10.0.2.1/24")},
		{ID: "host-b:eth0", Node: "host-b", Address: network.MustParsePrefix("10.0.2.20/24")},
	}
	for _, iface := range ifaces {
		if err := topo.AddInterface(iface); err != nil {
			t.Fatalf("AddInterface %s: %v", iface.ID, err)
		}
	}

	links := []*topology.Link{
		{ID: "link-a", A: "host-a:eth0", B: "router-r1:eth0", Delay: sim.Ms(1), Up: true},
		{ID: "link-b", A: "router-r1:eth1", B: "host-b:eth0", Delay: sim.Ms(1), Up: true},
	}
	for _, link := range links {
		if err := topo.AddLink(link); err != nil {
			t.Fatalf("AddLink %s: %v", link.ID, err)
		}
	}
	return topo
}

func pingAction() Action {
	return Action{
		At:   0,
		Kind: ActionSendPing,
		From: "host-a",
		To:   network.MustParseIPv4("10.0.2.20"),
	}
}

func newTestEngine(t *testing.T, topo *topology.Topology, opts Options) *Engine {
	t.Helper()
	e, err := New(topo, opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return e
}

func runEngine(t *testing.T, e *Engine) Result {
	t.Helper()
	if err := e.RunToCompletion(); err != nil {
		t.Fatalf("RunToCompletion: %v", err)
	}
	return e.Result()
}

func recordKinds(records []eventlog.Record) []string {
	kinds := make([]string, len(records))
	for i, rec := range records {
		kinds[i] = string(rec.Kind)
	}
	return kinds
}

func findRecord(records []eventlog.Record, kind sim.EventKind) (eventlog.Record, bool) {
	for _, rec := range records {
		if rec.Kind == kind {
			return rec, true
		}
	}
	return eventlog.Record{}, false
}

func TestSuccessfulPing(t *testing.T) {
	e := newTestEngine(t, newTestTopology(t), Options{Actions: []Action{pingAction()}})
	result := runEngine(t, e)

	if !result.Completed {
		t.Fatal("run did not complete")
	}
	if result.Delivered != 2 {
		t.Fatalf("Delivered = %d, want 2 (request and reply)", result.Delivered)
	}
	if result.Dropped != 0 {
		t.Fatalf("Dropped = %d, want 0", result.Dropped)
	}
	if result.Duration != sim.Ms(4) {
		t.Fatalf("Duration = %s, want 4ms", result.Duration)
	}

	records := e.Records()
	reply, ok := findRecord(records, sim.PingReply)
	if !ok {
		t.Fatal("no ping.reply record")
	}
	if got := reply.Details["rttMs"]; got != "4.000" {
		t.Fatalf("rttMs = %q, want 4.000", got)
	}
	if got := reply.Details["request"]; got != "p-001" {
		t.Fatalf("request = %q, want p-001", got)
	}
	if got := reply.Details["reply"]; got != "p-002" {
		t.Fatalf("reply = %q, want p-002", got)
	}
}

func TestPingTraceOrder(t *testing.T) {
	e := newTestEngine(t, newTestTopology(t), Options{Actions: []Action{pingAction()}})
	runEngine(t, e)

	want := []string{
		"run.started",
		"packet.created",
		"route.selected",
		"link.transmit",
		"link.delivered",
		"packet.received",
		"route.selected",
		"packet.forwarded",
		"link.transmit",
		"link.delivered",
		"packet.received",
		"packet.delivered",
		"packet.created",
		"route.selected",
		"link.transmit",
		"link.delivered",
		"packet.received",
		"route.selected",
		"packet.forwarded",
		"link.transmit",
		"link.delivered",
		"packet.received",
		"packet.delivered",
		"ping.reply",
		"run.completed",
	}
	got := recordKinds(e.Records())
	if len(got) != len(want) {
		t.Fatalf("got %d records, want %d:\n%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("record %d = %q, want %q\nfull trace: %v", i, got[i], want[i], got)
		}
	}
}

func TestRecordTimesAreMonotonic(t *testing.T) {
	e := newTestEngine(t, newTestTopology(t), Options{Actions: []Action{pingAction()}})
	runEngine(t, e)

	var last sim.SimTime
	for _, rec := range e.Records() {
		if rec.At < last {
			t.Fatalf("record %d moved time backward: %s after %s", rec.Index, rec.At, last)
		}
		last = rec.At
	}
}

func TestPingPacketIdentityIsStable(t *testing.T) {
	e := newTestEngine(t, newTestTopology(t), Options{Actions: []Action{pingAction()}})
	runEngine(t, e)

	for _, rec := range e.Records() {
		if rec.PacketID != "" && rec.PacketID != "p-001" && rec.PacketID != "p-002" {
			t.Fatalf("unexpected packet ID %q", rec.PacketID)
		}
	}
}
