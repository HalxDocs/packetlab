// Package trace assembles a self-contained run bundle for the web viewer:
// the scenario, a renderable topology, the event log, and the result summary.
package trace

import (
	"github.com/HalxDocs/packetlab/internal/engine"
	"github.com/HalxDocs/packetlab/internal/eventlog"
	"github.com/HalxDocs/packetlab/internal/scenario"
	"github.com/HalxDocs/packetlab/internal/topology"
)

// SchemaVersion is the bundle format understood by the web viewer.
const SchemaVersion = "1.0"

// Bundle is a complete trace document.
type Bundle struct {
	SchemaVersion string            `json:"schemaVersion"`
	Scenario      scenario.Scenario `json:"scenario"`
	Topology      Topology          `json:"topology"`
	Records       []eventlog.Record `json:"records"`
	Result        Result            `json:"result"`
}

// Topology is the renderable form of a topology.
type Topology struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

// Node is a host or router with its interfaces.
type Node struct {
	ID         string      `json:"id"`
	Name       string      `json:"name,omitempty"`
	Kind       string      `json:"kind"`
	Down       bool        `json:"down,omitempty"`
	Interfaces []Interface `json:"interfaces"`
	Routes     []Route     `json:"routes,omitempty"`
}

// Route is a static route entry shown in the node inspector.
type Route struct {
	Network   string `json:"network"`
	NextHop   string `json:"nextHop,omitempty"`
	Interface string `json:"interface"`
	Metric    int    `json:"metric,omitempty"`
}

// Interface is an addressed attachment point.
type Interface struct {
	ID      string `json:"id"`
	Node    string `json:"node"`
	Address string `json:"address"`
	Subnet  string `json:"subnet"`
}

// Link is a point-to-point connection.
type Link struct {
	ID           string  `json:"id"`
	A            string  `json:"a"`
	B            string  `json:"b"`
	DelayMs      float64 `json:"delayMs"`
	BandwidthBps uint64  `json:"bandwidthBps,omitempty"`
	Up           bool    `json:"up"`
}

// Result is the engine's run summary.
type Result struct {
	Processed    uint64  `json:"processed"`
	Delivered    int     `json:"delivered"`
	Dropped      int     `json:"dropped"`
	DurationMs   float64 `json:"durationMs"`
	Completed    bool    `json:"completed"`
	LimitReached bool    `json:"limitReached"`
	Records      int     `json:"records"`
}

// Build converts engine output into a bundle.
func Build(s *scenario.Scenario, topo *topology.Topology, records []eventlog.Record, result engine.Result) Bundle {
	return Bundle{
		SchemaVersion: SchemaVersion,
		Scenario:      *s,
		Topology:      buildTopology(topo),
		Records:       records,
		Result: Result{
			Processed:    result.Processed,
			Delivered:    result.Delivered,
			Dropped:      result.Dropped,
			DurationMs:   result.Duration.Milliseconds(),
			Completed:    result.Completed,
			LimitReached: result.LimitReached,
			Records:      len(records),
		},
	}
}

func buildTopology(topo *topology.Topology) Topology {
	out := Topology{
		Nodes: make([]Node, 0, topo.NodeCount()),
		Links: make([]Link, 0, topo.LinkCount()),
	}

	for _, node := range topo.Nodes() {
		n := Node{
			ID:   string(node.ID),
			Name: node.Name,
			Kind: string(node.Kind),
			Down: node.Down,
		}
		for _, iface := range topo.Interfaces() {
			if iface.Node != node.ID {
				continue
			}
			n.Interfaces = append(n.Interfaces, Interface{
				ID:      string(iface.ID),
				Node:    string(iface.Node),
				Address: iface.Address.String(),
				Subnet:  iface.Subnet().String(),
			})
		}
		for _, route := range node.Routes {
			r := Route{
				Network:   route.Network.String(),
				Interface: route.Interface,
				Metric:    route.Metric,
			}
			if route.NextHop.IsValid() && !route.NextHop.IsUnspecified() {
				r.NextHop = route.NextHop.String()
			}
			n.Routes = append(n.Routes, r)
		}
		out.Nodes = append(out.Nodes, n)
	}

	for _, link := range topo.Links() {
		out.Links = append(out.Links, Link{
			ID:           string(link.ID),
			A:            string(link.A),
			B:            string(link.B),
			DelayMs:      link.Delay.Milliseconds(),
			BandwidthBps: link.BandwidthBps,
			Up:           link.Up,
		})
	}
	return out
}
