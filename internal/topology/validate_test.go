package topology

import (
	"strings"
	"testing"

	"github.com/HalxDocs/packetlab/internal/network"
	"github.com/HalxDocs/packetlab/internal/sim"
)

func TestValidateAcceptsCanonicalTopology(t *testing.T) {
	if err := addTestTopology(t).Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidateRejectsEmptyTopology(t *testing.T) {
	err := New().Validate()
	if err == nil || !strings.Contains(err.Error(), "at least one node") {
		t.Fatalf("err = %v, want empty topology problem", err)
	}
}

func TestValidateRejectsUnknownKind(t *testing.T) {
	topo := New()
	if err := topo.AddNode(&Node{ID: "n1", Kind: "switch"}); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	if err := topo.AddInterface(&Interface{ID: "n1:eth0", Node: "n1", Address: network.MustParsePrefix("10.0.0.1/24")}); err != nil {
		t.Fatalf("AddInterface: %v", err)
	}
	err := topo.Validate()
	if err == nil || !strings.Contains(err.Error(), "unknown node kind") {
		t.Fatalf("err = %v, want unknown kind problem", err)
	}
}

func TestValidateRejectsNodeWithoutInterfaces(t *testing.T) {
	topo := New()
	if err := topo.AddNode(&Node{ID: "n1", Kind: KindHost}); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	err := topo.Validate()
	if err == nil || !strings.Contains(err.Error(), "no interfaces") {
		t.Fatalf("err = %v, want missing interface problem", err)
	}
}

func TestValidateRejectsDuplicateAddress(t *testing.T) {
	topo := addTestTopology(t)
	if err := topo.AddNode(&Node{ID: "host-c", Kind: KindHost}); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	if err := topo.AddInterface(&Interface{
		ID:      "host-c:eth0",
		Node:    "host-c",
		Address: network.MustParsePrefix("10.0.2.20/24"),
	}); err != nil {
		t.Fatalf("AddInterface: %v", err)
	}
	err := topo.Validate()
	if err == nil || !strings.Contains(err.Error(), "duplicate address") {
		t.Fatalf("err = %v, want duplicate address problem", err)
	}
}

func TestValidateRejectsRouteWithUnknownInterface(t *testing.T) {
	topo := addTestTopology(t)
	node, _ := topo.Node("host-a")
	node.Routes = []network.Route{{
		Network:   network.MustParsePrefix("0.0.0.0/0"),
		NextHop:   network.MustParseIPv4("10.0.1.1"),
		Interface: "host-a:eth9",
	}}
	err := topo.Validate()
	if err == nil || !strings.Contains(err.Error(), "unknown interface") {
		t.Fatalf("err = %v, want unknown route interface problem", err)
	}
}

func TestValidateRejectsRouteOnWrongNode(t *testing.T) {
	topo := addTestTopology(t)
	node, _ := topo.Node("host-a")
	node.Routes = []network.Route{{
		Network:   network.MustParsePrefix("0.0.0.0/0"),
		NextHop:   network.MustParseIPv4("10.0.1.1"),
		Interface: "host-b:eth0",
	}}
	err := topo.Validate()
	if err == nil || !strings.Contains(err.Error(), "belongs to node") {
		t.Fatalf("err = %v, want wrong-node route problem", err)
	}
}

func TestValidateRejectsNegativeLinkValues(t *testing.T) {
	topo := addTestTopology(t)
	link, _ := topo.Link("link-a")
	link.Delay = -sim.Millisecond
	link.QueueLimit = -1
	err := topo.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "delay must not be negative") {
		t.Errorf("missing delay problem: %s", msg)
	}
	if !strings.Contains(msg, "queue limit must not be negative") {
		t.Errorf("missing queue limit problem: %s", msg)
	}
}

func TestValidateProblemsAreSorted(t *testing.T) {
	topo := New()
	if err := topo.AddNode(&Node{ID: "b", Kind: "bad"}); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	if err := topo.AddNode(&Node{ID: "a", Kind: "bad"}); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	err := topo.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	msg := err.Error()
	if strings.Index(msg, "nodes[a]") > strings.Index(msg, "nodes[b]") {
		t.Fatalf("problems are not sorted: %s", msg)
	}
}
