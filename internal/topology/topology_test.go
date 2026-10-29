package topology

import (
	"testing"

	"github.com/HalxDocs/packetlab/internal/network"
	"github.com/HalxDocs/packetlab/internal/sim"
)

func addTestTopology(t *testing.T) *Topology {
	t.Helper()
	topo := New()

	if err := topo.AddNode(&Node{ID: "host-a", Kind: KindHost}); err != nil {
		t.Fatalf("AddNode host-a: %v", err)
	}
	if err := topo.AddNode(&Node{ID: "router-r1", Kind: KindRouter}); err != nil {
		t.Fatalf("AddNode router-r1: %v", err)
	}
	if err := topo.AddNode(&Node{ID: "host-b", Kind: KindHost}); err != nil {
		t.Fatalf("AddNode host-b: %v", err)
	}

	ifaces := []*Interface{
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

	links := []*Link{
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

func TestAddNodeRejectsDuplicate(t *testing.T) {
	topo := New()
	if err := topo.AddNode(&Node{ID: "n1", Kind: KindHost}); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	if err := topo.AddNode(&Node{ID: "n1", Kind: KindHost}); err == nil {
		t.Fatal("expected duplicate node error")
	}
}

func TestAddNodeRejectsEmptyID(t *testing.T) {
	if err := New().AddNode(&Node{Kind: KindHost}); err == nil {
		t.Fatal("expected empty node ID error")
	}
}

func TestAddInterfaceRequiresNode(t *testing.T) {
	topo := New()
	err := topo.AddInterface(&Interface{ID: "if1", Node: "missing"})
	if err == nil {
		t.Fatal("expected unknown node error")
	}
}

func TestAddInterfaceUpdatesNode(t *testing.T) {
	topo := New()
	if err := topo.AddNode(&Node{ID: "n1", Kind: KindHost}); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	if err := topo.AddInterface(&Interface{ID: "if1", Node: "n1"}); err != nil {
		t.Fatalf("AddInterface: %v", err)
	}
	node, _ := topo.Node("n1")
	if !node.HasInterface("if1") {
		t.Fatal("node does not list its interface")
	}
}

func TestAddLinkRejectsUnknownInterface(t *testing.T) {
	topo := addTestTopology(t)
	err := topo.AddLink(&Link{ID: "link-x", A: "host-a:eth0", B: "missing:eth0", Up: true})
	if err == nil {
		t.Fatal("expected unknown interface error")
	}
}

func TestAddLinkRejectsSelfConnection(t *testing.T) {
	topo := New()
	if err := topo.AddNode(&Node{ID: "n1", Kind: KindHost}); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	if err := topo.AddInterface(&Interface{ID: "n1:eth0", Node: "n1"}); err != nil {
		t.Fatalf("AddInterface: %v", err)
	}
	if err := topo.AddInterface(&Interface{ID: "n1:eth1", Node: "n1"}); err != nil {
		t.Fatalf("AddInterface: %v", err)
	}
	if err := topo.AddLink(&Link{ID: "loop", A: "n1:eth0", B: "n1:eth1", Up: true}); err == nil {
		t.Fatal("expected self-connection error")
	}
}

func TestAddLinkRejectsReusedInterface(t *testing.T) {
	topo := addTestTopology(t)
	err := topo.AddLink(&Link{ID: "link-c", A: "host-a:eth0", B: "host-b:eth0", Up: true})
	if err == nil {
		t.Fatal("expected interface already attached error")
	}
}

func TestLookups(t *testing.T) {
	topo := addTestTopology(t)

	if _, ok := topo.Node("router-r1"); !ok {
		t.Fatal("Node(router-r1) not found")
	}
	if _, ok := topo.Interface("host-b:eth0"); !ok {
		t.Fatal("Interface(host-b:eth0) not found")
	}
	link, ok := topo.LinkForInterface("router-r1:eth0")
	if !ok || link.ID != "link-a" {
		t.Fatalf("LinkForInterface = %+v, %v", link, ok)
	}
	peer, ok := link.Peer("router-r1:eth0")
	if !ok || peer != "host-a:eth0" {
		t.Fatalf("Peer = %q, %v", peer, ok)
	}
}

func TestNodeByAddress(t *testing.T) {
	topo := addTestTopology(t)
	node, iface, ok := topo.NodeByAddress(network.MustParseIPv4("10.0.2.20"))
	if !ok {
		t.Fatal("address not found")
	}
	if node.ID != "host-b" || iface.ID != "host-b:eth0" {
		t.Fatalf("got node %q iface %q", node.ID, iface.ID)
	}
	if _, _, ok := topo.NodeByAddress(network.MustParseIPv4("10.9.9.9")); ok {
		t.Fatal("unexpected match")
	}
}

func TestSortedCollections(t *testing.T) {
	topo := addTestTopology(t)

	nodes := topo.Nodes()
	wantNodes := []NodeID{"host-a", "host-b", "router-r1"}
	for i, want := range wantNodes {
		if nodes[i].ID != want {
			t.Fatalf("node %d = %q, want %q", i, nodes[i].ID, want)
		}
	}

	links := topo.Links()
	if links[0].ID != "link-a" || links[1].ID != "link-b" {
		t.Fatalf("links not sorted: %q, %q", links[0].ID, links[1].ID)
	}

	if got := topo.NodeCount(); got != 3 {
		t.Fatalf("NodeCount = %d, want 3", got)
	}
	if got := topo.LinkCount(); got != 2 {
		t.Fatalf("LinkCount = %d, want 2", got)
	}
}
