package scenario

import (
	"fmt"

	"github.com/HalxDocs/packetlab/internal/engine"
	"github.com/HalxDocs/packetlab/internal/network"
	"github.com/HalxDocs/packetlab/internal/sim"
	"github.com/HalxDocs/packetlab/internal/topology"
)

// Build converts a validated scenario into a topology and engine options.
func (s *Scenario) Build() (*topology.Topology, engine.Options, error) {
	if err := s.Validate(); err != nil {
		return nil, engine.Options{}, err
	}

	topo := topology.New()

	for _, n := range s.Nodes {
		node := &topology.Node{
			ID:   topology.NodeID(n.ID),
			Name: n.Name,
			Kind: topology.NodeKind(n.Kind),
			Down: n.Down,
		}
		if err := topo.AddNode(node); err != nil {
			return nil, engine.Options{}, err
		}
	}

	for _, n := range s.Nodes {
		node, _ := topo.Node(topology.NodeID(n.ID))
		for _, iface := range n.Interfaces {
			prefix, err := network.ParsePrefix(iface.Address)
			if err != nil {
				return nil, engine.Options{}, err
			}
			if err := topo.AddInterface(&topology.Interface{
				ID:      topology.InterfaceID(iface.ID),
				Node:    node.ID,
				Address: prefix,
			}); err != nil {
				return nil, engine.Options{}, err
			}
		}
		for _, r := range n.Routes {
			route, err := buildRoute(r)
			if err != nil {
				return nil, engine.Options{}, err
			}
			node.Routes = append(node.Routes, route)
		}
	}

	for _, l := range s.Links {
		link := &topology.Link{
			ID:           topology.LinkID(l.ID),
			A:            topology.InterfaceID(l.A),
			B:            topology.InterfaceID(l.B),
			Delay:        sim.Ms(l.DelayMs),
			BandwidthBps: l.BandwidthBps,
			QueueLimit:   l.QueueLimit,
			Up:           !l.Down,
		}
		if err := topo.AddLink(link); err != nil {
			return nil, engine.Options{}, err
		}
	}

	if err := topo.Validate(); err != nil {
		return nil, engine.Options{}, err
	}

	actions := make([]engine.Action, 0, len(s.Actions))
	for _, a := range s.Actions {
		to, err := network.ParseIPv4(a.To)
		if err != nil {
			return nil, engine.Options{}, err
		}
		actions = append(actions, engine.Action{
			At:       sim.Ms(a.AtMs),
			Kind:     engine.ActionKind(a.Kind),
			From:     topology.NodeID(a.From),
			To:       to,
			Count:    a.Count,
			Interval: sim.Ms(a.IntervalMs),
			TTL:      a.TTL,
		})
	}

	return topo, engine.Options{Seed: s.Seed, Actions: actions}, nil
}

func buildRoute(r Route) (network.Route, error) {
	prefix, err := network.ParsePrefix(r.Network)
	if err != nil {
		return network.Route{}, err
	}
	route := network.Route{
		Network:   prefix.Masked(),
		Interface: r.Interface,
		Metric:    r.Metric,
	}
	if r.NextHop != "" {
		nextHop, err := network.ParseIPv4(r.NextHop)
		if err != nil {
			return network.Route{}, err
		}
		route.NextHop = nextHop
	}
	return route, nil
}

// Summary returns a one-line description for CLI headers.
func (s *Scenario) Summary() string {
	return fmt.Sprintf("%s (schema %s, seed %d)", s.ID, s.SchemaVersion, s.Seed)
}
