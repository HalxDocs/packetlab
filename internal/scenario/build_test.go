package scenario

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HalxDocs/packetlab/internal/network"
	"github.com/HalxDocs/packetlab/internal/sim"
	"github.com/HalxDocs/packetlab/internal/topology"
)

func TestBuildCanonicalTopology(t *testing.T) {
	s, err := Load(scenarioPath("two-hosts-one-router.json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	topo, opts, err := s.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if topo.NodeCount() != 3 || topo.LinkCount() != 2 {
		t.Fatalf("topology = %d nodes, %d links", topo.NodeCount(), topo.LinkCount())
	}
	link, ok := topo.Link("link-a")
	if !ok || link.Delay != sim.Ms(1) || !link.Up {
		t.Fatalf("link-a = %+v, %v", link, ok)
	}
	host, _ := topo.Node("host-a")
	if len(host.Routes) != 1 || !host.Routes[0].Network.IsDefault() {
		t.Fatalf("host-a routes = %+v", host.Routes)
	}
	if len(opts.Actions) != 1 || opts.Actions[0].Kind != "send_ping" {
		t.Fatalf("actions = %+v", opts.Actions)
	}
	if opts.Seed != 42 {
		t.Fatalf("seed = %d, want 42", opts.Seed)
	}
}

func TestBuildMasksRouteNetworks(t *testing.T) {
	s, err := Decode(strings.NewReader(validJSON()))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	s.Nodes[0].Routes = []Route{{
		Network:   "10.0.0.99/24",
		NextHop:   "10.0.0.2",
		Interface: "a:eth0",
	}}
	topo, _, err := s.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	node, _ := topo.Node("a")
	if got := node.Routes[0].Network.String(); got != "10.0.0.0/24" {
		t.Fatalf("route network = %q, want masked 10.0.0.0/24", got)
	}
}

func TestBuildDownLink(t *testing.T) {
	s, err := Load(scenarioPath("link-down.json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	topo, _, err := s.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	link, _ := topo.Link("link-b")
	if link.Up {
		t.Fatal("link-b should be down")
	}
}

func TestBuildRejectsInvalidScenario(t *testing.T) {
	s, err := Decode(strings.NewReader(validJSON()))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	s.Nodes[0].Interfaces[0].Address = "bad"
	if _, _, err := s.Build(); err == nil {
		t.Fatal("expected build error")
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	s, err := Load(scenarioPath("two-hosts-one-router.json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	data, err := Marshal(s)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	again, err := Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if again.ID != s.ID || len(again.Nodes) != len(s.Nodes) {
		t.Fatalf("round trip mismatch: %+v", again)
	}
}

func TestBuildRouteParsing(t *testing.T) {
	route, err := buildRoute(Route{Network: "10.0.0.0/8", NextHop: "10.1.2.3", Interface: "eth0", Metric: 5})
	if err != nil {
		t.Fatalf("buildRoute: %v", err)
	}
	if route.NextHop.String() != "10.1.2.3" || route.Metric != 5 {
		t.Fatalf("route = %+v", route)
	}
	if route.Network.String() != "10.0.0.0/8" {
		t.Fatalf("network = %s", route.Network)
	}
}

func TestBuildDirectRouteHasNoNextHop(t *testing.T) {
	route, err := buildRoute(Route{Network: "10.0.0.0/24", Interface: "eth0"})
	if err != nil {
		t.Fatalf("buildRoute: %v", err)
	}
	if !route.DirectlyConnected() {
		t.Fatal("route without a next hop should be directly connected")
	}
	if route.NextHop != network.IPv4Zero {
		t.Fatalf("next hop = %v", route.NextHop)
	}
}

func TestSummary(t *testing.T) {
	s, _ := Load(scenarioPath("two-hosts-one-router.json"))
	if got := s.Summary(); !strings.Contains(got, "two-hosts-one-router") || !strings.Contains(got, "seed 42") {
		t.Fatalf("Summary = %q", got)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join("..", "..", "scenarios", "nope.json")); err == nil {
		t.Fatal("expected open error")
	}
}

func TestTopologyKindMatchesScenario(t *testing.T) {
	s, _ := Load(scenarioPath("two-hosts-one-router.json"))
	topo, _, err := s.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	node, _ := topo.Node("router-r1")
	if node.Kind != topology.KindRouter {
		t.Fatalf("kind = %q", node.Kind)
	}
}
