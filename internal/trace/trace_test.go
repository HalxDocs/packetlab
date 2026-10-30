package trace

import (
	"encoding/json"
	"testing"

	"github.com/HalxDocs/packetlab/internal/engine"
	"github.com/HalxDocs/packetlab/internal/scenario"
)

func buildTestBundle(t *testing.T) Bundle {
	t.Helper()
	s, err := scenario.Load("../../scenarios/two-hosts-one-router.json")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	topo, opts, err := s.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	e, err := engine.New(topo, opts)
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	if err := e.RunToCompletion(); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return Build(s, topo, e.Records(), e.Result())
}

func TestBuildBundle(t *testing.T) {
	b := buildTestBundle(t)
	if b.SchemaVersion != SchemaVersion {
		t.Fatalf("schema = %q", b.SchemaVersion)
	}
	if len(b.Topology.Nodes) != 3 {
		t.Fatalf("nodes = %d, want 3", len(b.Topology.Nodes))
	}
	if len(b.Topology.Links) != 2 {
		t.Fatalf("links = %d, want 2", len(b.Topology.Links))
	}
	if len(b.Records) != 25 {
		t.Fatalf("records = %d, want 25", len(b.Records))
	}
	if b.Result.Delivered != 2 || !b.Result.Completed {
		t.Fatalf("result = %+v", b.Result)
	}
	if b.Result.Records != 25 {
		t.Fatalf("result records = %d, want 25", b.Result.Records)
	}
}

func TestBundleJSONRoundTrip(t *testing.T) {
	b := buildTestBundle(t)
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out Bundle
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.Records) != len(b.Records) || len(out.Topology.Nodes) != 3 {
		t.Fatalf("round trip mismatch")
	}
}

func TestNodeInterfacesAreNested(t *testing.T) {
	b := buildTestBundle(t)
	for _, node := range b.Topology.Nodes {
		if node.ID == "router-r1" {
			if len(node.Interfaces) != 2 {
				t.Fatalf("router interfaces = %d, want 2", len(node.Interfaces))
			}
			if node.Interfaces[0].Subnet != "10.0.1.0/24" {
				t.Fatalf("subnet = %q", node.Interfaces[0].Subnet)
			}
		}
	}
}

func TestNodeRoutesAreProjected(t *testing.T) {
	b := buildTestBundle(t)
	var host *Node
	for i := range b.Topology.Nodes {
		if b.Topology.Nodes[i].ID == "host-a" {
			host = &b.Topology.Nodes[i]
			break
		}
	}
	if host == nil {
		t.Fatal("host-a missing from topology")
	}
	if len(host.Routes) != 1 {
		t.Fatalf("host-a routes = %d, want 1", len(host.Routes))
	}
	route := host.Routes[0]
	if route.Network != "0.0.0.0/0" {
		t.Errorf("network = %q", route.Network)
	}
	if route.NextHop != "10.0.1.1" {
		t.Errorf("next hop = %q", route.NextHop)
	}
	if route.Interface != "host-a:eth0" {
		t.Errorf("interface = %q", route.Interface)
	}
}

func TestDirectRouteOmitsNextHop(t *testing.T) {
	b := buildTestBundle(t)
	var router *Node
	for i := range b.Topology.Nodes {
		if b.Topology.Nodes[i].ID == "router-r1" {
			router = &b.Topology.Nodes[i]
			break
		}
	}
	if router == nil {
		t.Fatal("router-r1 missing from topology")
	}
	for _, route := range router.Routes {
		if route.Network == "10.0.2.0/24" {
			if route.NextHop != "" {
				t.Fatalf("direct route next hop = %q, want omitted", route.NextHop)
			}
			return
		}
	}
	t.Fatal("router-r1 is missing its direct route to 10.0.2.0/24")
}

func TestLinkDelay(t *testing.T) {
	b := buildTestBundle(t)
	for _, link := range b.Topology.Links {
		if link.DelayMs != 1 {
			t.Fatalf("link %s delay = %v, want 1", link.ID, link.DelayMs)
		}
		if !link.Up {
			t.Fatalf("link %s should be up", link.ID)
		}
	}
}

func TestRecordsSurviveEncoding(t *testing.T) {
	b := buildTestBundle(t)
	data, err := json.Marshal(b.Records)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var records []map[string]any
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(records) != 25 {
		t.Fatalf("records = %d", len(records))
	}
	if _, ok := records[0]["atMs"]; !ok {
		t.Fatalf("record 0 missing atMs: %v", records[0])
	}
}
