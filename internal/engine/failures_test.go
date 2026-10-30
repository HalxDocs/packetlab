package engine

import (
	"testing"

	"github.com/HalxDocs/packetlab/internal/network"
	"github.com/HalxDocs/packetlab/internal/sim"
	"github.com/HalxDocs/packetlab/internal/topology"
)

func TestMissingRoute(t *testing.T) {
	topo := newTestTopology(t)
	host, _ := topo.Node("host-a")
	host.Routes = nil

	e := newTestEngine(t, topo, Options{Actions: []Action{pingAction()}})
	result := runEngine(t, e)

	if result.Dropped != 1 {
		t.Fatalf("Dropped = %d, want 1", result.Dropped)
	}
	if result.Delivered != 0 {
		t.Fatalf("Delivered = %d, want 0", result.Delivered)
	}
	drop, ok := findRecord(e.Records(), sim.PacketDropped)
	if !ok {
		t.Fatal("no packet.dropped record")
	}
	if drop.Reason != sim.ReasonNoRoute {
		t.Fatalf("Reason = %q, want %q", drop.Reason, sim.ReasonNoRoute)
	}
	if drop.At != 0 {
		t.Fatalf("drop at %s, want 0ms", drop.At)
	}
	if drop.NodeID != "host-a" {
		t.Fatalf("drop node = %q, want host-a", drop.NodeID)
	}
}

func TestLinkDown(t *testing.T) {
	topo := newTestTopology(t)
	link, _ := topo.Link("link-b")
	link.Up = false

	e := newTestEngine(t, topo, Options{Actions: []Action{pingAction()}})
	result := runEngine(t, e)

	if result.Dropped != 1 {
		t.Fatalf("Dropped = %d, want 1", result.Dropped)
	}
	drop, ok := findRecord(e.Records(), sim.PacketDropped)
	if !ok {
		t.Fatal("no packet.dropped record")
	}
	if drop.Reason != sim.ReasonLinkDown {
		t.Fatalf("Reason = %q, want %q", drop.Reason, sim.ReasonLinkDown)
	}
	if drop.LinkID != "link-b" {
		t.Fatalf("LinkID = %q, want link-b", drop.LinkID)
	}
	if drop.NodeID != "router-r1" {
		t.Fatalf("drop node = %q, want router-r1", drop.NodeID)
	}
	if _, ok := findRecord(e.Records(), sim.PacketDelivered); ok {
		t.Fatal("packet should not be delivered on a down link")
	}
}

func TestTTLExpiry(t *testing.T) {
	action := pingAction()
	action.TTL = 1

	e := newTestEngine(t, newTestTopology(t), Options{Actions: []Action{action}})
	result := runEngine(t, e)

	if result.Dropped != 1 {
		t.Fatalf("Dropped = %d, want 1", result.Dropped)
	}
	drop, ok := findRecord(e.Records(), sim.PacketDropped)
	if !ok {
		t.Fatal("no packet.dropped record")
	}
	if drop.Reason != sim.ReasonTTLExpired {
		t.Fatalf("Reason = %q, want %q", drop.Reason, sim.ReasonTTLExpired)
	}
	if drop.NodeID != "router-r1" {
		t.Fatalf("drop node = %q, want router-r1", drop.NodeID)
	}
	if drop.At != sim.Ms(1) {
		t.Fatalf("drop at %s, want 1ms", drop.At)
	}
}

// newLoopTopology builds host-a — r1 — r2 with routes that point at each
// other for 10.9.9.9.
func newLoopTopology(t *testing.T) *topology.Topology {
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
				{
					Network:   network.MustParsePrefix("10.9.9.9/32"),
					NextHop:   network.MustParseIPv4("10.0.2.2"),
					Interface: "router-r1:eth1",
				},
			},
		},
		{
			ID: "router-r2", Kind: topology.KindRouter,
			Routes: []network.Route{
				{Network: network.MustParsePrefix("10.0.2.0/24"), Interface: "router-r2:eth0"},
				{
					Network:   network.MustParsePrefix("10.9.9.9/32"),
					NextHop:   network.MustParseIPv4("10.0.2.1"),
					Interface: "router-r2:eth0",
				},
			},
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
		{ID: "router-r2:eth0", Node: "router-r2", Address: network.MustParsePrefix("10.0.2.2/24")},
	}
	for _, iface := range ifaces {
		if err := topo.AddInterface(iface); err != nil {
			t.Fatalf("AddInterface %s: %v", iface.ID, err)
		}
	}

	links := []*topology.Link{
		{ID: "link-a", A: "host-a:eth0", B: "router-r1:eth0", Delay: sim.Ms(1), Up: true},
		{ID: "link-b", A: "router-r1:eth1", B: "router-r2:eth0", Delay: sim.Ms(1), Up: true},
	}
	for _, link := range links {
		if err := topo.AddLink(link); err != nil {
			t.Fatalf("AddLink %s: %v", link.ID, err)
		}
	}
	return topo
}

func TestRoutingLoopExpiresTTL(t *testing.T) {
	action := Action{
		At:   0,
		Kind: ActionSendPing,
		From: "host-a",
		To:   network.MustParseIPv4("10.9.9.9"),
		TTL:  3,
	}
	e := newTestEngine(t, newLoopTopology(t), Options{Actions: []Action{action}})
	result := runEngine(t, e)

	if result.Dropped != 1 {
		t.Fatalf("Dropped = %d, want 1", result.Dropped)
	}
	drop, ok := findRecord(e.Records(), sim.PacketDropped)
	if !ok {
		t.Fatal("no packet.dropped record")
	}
	if drop.Reason != sim.ReasonTTLExpired {
		t.Fatalf("Reason = %q, want ttl_expired", drop.Reason)
	}
	if drop.At != sim.Ms(3) {
		t.Fatalf("drop at %s, want 3ms", drop.At)
	}
}

func TestUnknownDestination(t *testing.T) {
	topo := newTestTopology(t)
	iface, _ := topo.Interface("host-b:eth0")
	iface.Address = network.MustParsePrefix("10.0.2.99/24")

	e := newTestEngine(t, topo, Options{Actions: []Action{pingAction()}})
	result := runEngine(t, e)

	drop, ok := findRecord(e.Records(), sim.PacketDropped)
	if !ok {
		t.Fatal("no packet.dropped record")
	}
	if drop.Reason != sim.ReasonUnknownDestination {
		t.Fatalf("Reason = %q, want %q", drop.Reason, sim.ReasonUnknownDestination)
	}
	if result.Dropped != 1 {
		t.Fatalf("Dropped = %d, want 1", result.Dropped)
	}
}

func TestNodeDown(t *testing.T) {
	topo := newTestTopology(t)
	node, _ := topo.Node("host-b")
	node.Down = true

	e := newTestEngine(t, topo, Options{Actions: []Action{pingAction()}})
	runEngine(t, e)

	drop, ok := findRecord(e.Records(), sim.PacketDropped)
	if !ok {
		t.Fatal("no packet.dropped record")
	}
	if drop.Reason != sim.ReasonNodeDown {
		t.Fatalf("Reason = %q, want %q", drop.Reason, sim.ReasonNodeDown)
	}
}

func TestNoActionsCompletes(t *testing.T) {
	e := newTestEngine(t, newTestTopology(t), Options{})
	result := runEngine(t, e)

	if !result.Completed {
		t.Fatal("run with no actions did not complete")
	}
	if result.Processed != 0 {
		t.Fatalf("Processed = %d, want 0", result.Processed)
	}
	if got := len(e.Records()); got != 2 {
		t.Fatalf("records = %d, want run.started and run.completed", got)
	}
}
