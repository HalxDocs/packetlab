package scenario

import (
	"fmt"
	"sort"
	"strings"

	"github.com/HalxDocs/packetlab/internal/network"
)

// Limits protect the engine from pathological imported files.
const (
	MaxNodes   = 256
	MaxLinks   = 1024
	MaxActions = 128
)

// ValidationError lists every problem found in a scenario file.
type ValidationError struct {
	Problems []string
}

// Error renders the problems in sorted order.
func (e *ValidationError) Error() string {
	problems := append([]string(nil), e.Problems...)
	sort.Strings(problems)
	return "scenario validation failed: " + strings.Join(problems, "; ")
}

// Validate checks schema, references, and value ranges.
func (s *Scenario) Validate() error {
	var problems []string
	add := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	if s.SchemaVersion != SchemaVersion {
		add("schemaVersion must be %q, got %q", SchemaVersion, s.SchemaVersion)
	}
	if strings.TrimSpace(s.ID) == "" {
		add("id must not be empty")
	}
	if strings.TrimSpace(s.Name) == "" {
		add("name must not be empty")
	}
	if len(s.Nodes) == 0 {
		add("nodes must not be empty")
	}
	if len(s.Nodes) > MaxNodes {
		add("nodes exceeds the limit of %d", MaxNodes)
	}
	if len(s.Links) > MaxLinks {
		add("links exceeds the limit of %d", MaxLinks)
	}
	if len(s.Actions) > MaxActions {
		add("actions exceeds the limit of %d", MaxActions)
	}

	nodes := make(map[string]bool, len(s.Nodes))
	interfaces := make(map[string]string, len(s.Nodes)*2)
	addresses := make(map[string]string, len(s.Nodes)*2)

	for i, node := range s.Nodes {
		path := fmt.Sprintf("nodes[%d]", i)
		if strings.TrimSpace(node.ID) == "" {
			add("%s.id must not be empty", path)
		} else if nodes[node.ID] {
			add("%s.id %q is duplicated", path, node.ID)
		}
		nodes[node.ID] = true

		switch node.Kind {
		case "host", "router":
		default:
			add("%s.kind must be \"host\" or \"router\", got %q", path, node.Kind)
		}
		if len(node.Interfaces) == 0 {
			add("%s must declare at least one interface", path)
		}
		for j, iface := range node.Interfaces {
			ifacePath := fmt.Sprintf("%s.interfaces[%d]", path, j)
			if strings.TrimSpace(iface.ID) == "" {
				add("%s.id must not be empty", ifacePath)
			} else if owner, exists := interfaces[iface.ID]; exists {
				add("%s.id %q is already used by node %q", ifacePath, iface.ID, owner)
			} else {
				interfaces[iface.ID] = node.ID
			}
			prefix, err := network.ParsePrefix(iface.Address)
			if err != nil {
				add("%s.address: %v", ifacePath, err)
			} else if prefix.Addr().IsUnspecified() {
				add("%s.address must not be 0.0.0.0", ifacePath)
			} else if owner, exists := addresses[prefix.Addr().String()]; exists {
				add("%s.address %s is already used by node %q", ifacePath, prefix.Addr(), owner)
			} else {
				addresses[prefix.Addr().String()] = node.ID
			}
		}
		for j, route := range node.Routes {
			routePath := fmt.Sprintf("%s.routes[%d]", path, j)
			if _, err := network.ParsePrefix(route.Network); err != nil {
				add("%s.network: %v", routePath, err)
			}
			if route.NextHop != "" {
				if _, err := network.ParseIPv4(route.NextHop); err != nil {
					add("%s.nextHop: %v", routePath, err)
				}
			}
			if route.Interface == "" {
				add("%s.interface must not be empty", routePath)
			}
			if route.Metric < 0 {
				add("%s.metric must not be negative", routePath)
			}
		}
	}

	links := make(map[string]bool, len(s.Links))
	for i, link := range s.Links {
		path := fmt.Sprintf("links[%d]", i)
		if strings.TrimSpace(link.ID) == "" {
			add("%s.id must not be empty", path)
		} else if links[link.ID] {
			add("%s.id %q is duplicated", path, link.ID)
		}
		links[link.ID] = true
		if link.A == "" || link.B == "" {
			add("%s endpoints must not be empty", path)
		} else if link.A == link.B {
			add("%s connects interface %q to itself", path, link.A)
		}
		for _, endpoint := range []string{link.A, link.B} {
			if endpoint == "" {
				continue
			}
			if _, ok := interfaces[endpoint]; !ok {
				add("%s references unknown interface %q", path, endpoint)
			}
		}
		if link.DelayMs < 0 {
			add("%s.delayMs must not be negative", path)
		}
		if link.QueueLimit < 0 {
			add("%s.queueLimit must not be negative", path)
		}
	}

	for i, action := range s.Actions {
		path := fmt.Sprintf("actions[%d]", i)
		if action.Kind != "send_ping" {
			add("%s.kind must be \"send_ping\", got %q", path, action.Kind)
		}
		if action.AtMs < 0 {
			add("%s.atMs must not be negative", path)
		}
		if !nodes[action.From] {
			add("%s.from references unknown node %q", path, action.From)
		}
		if _, err := network.ParseIPv4(action.To); err != nil {
			add("%s.to: %v", path, err)
		}
		if action.Count < 0 || action.Count > 1000 {
			add("%s.count must be between 1 and 1000", path)
		}
		if action.IntervalMs < 0 {
			add("%s.intervalMs must not be negative", path)
		}
		if action.TTL == 0 {
			// zero means the engine default
		} else if action.TTL < 1 || action.TTL > 255 {
			add("%s.ttl must be between 1 and 255", path)
		}
	}

	if len(problems) > 0 {
		return &ValidationError{Problems: problems}
	}
	return nil
}
