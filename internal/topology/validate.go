package topology

import (
	"fmt"
	"sort"
	"strings"
)

// Problem is one structural or addressing issue found during validation.
type Problem struct {
	Path    string
	Message string
}

// ValidationError aggregates every problem found in a topology.
type ValidationError struct {
	Problems []Problem
}

// Error renders the problems in a stable order.
func (e *ValidationError) Error() string {
	problems := append([]Problem(nil), e.Problems...)
	sort.SliceStable(problems, func(i, j int) bool {
		if problems[i].Path != problems[j].Path {
			return problems[i].Path < problems[j].Path
		}
		return problems[i].Message < problems[j].Message
	})
	parts := make([]string, 0, len(problems))
	for _, p := range problems {
		parts = append(parts, p.Path+": "+p.Message)
	}
	return "topology validation failed: " + strings.Join(parts, "; ")
}

// Validate checks the topology for structural and addressing problems.
func (t *Topology) Validate() error {
	var problems []Problem
	add := func(path, format string, args ...any) {
		problems = append(problems, Problem{Path: path, Message: fmt.Sprintf(format, args...)})
	}

	if len(t.nodes) == 0 {
		add("topology", "must contain at least one node")
	}

	seenIPs := make(map[string]InterfaceID)

	for _, nodeID := range t.NodeIDs() {
		node := t.nodes[nodeID]
		path := fmt.Sprintf("nodes[%s]", nodeID)

		if !node.Kind.Valid() {
			add(path+".kind", "unknown node kind %q", node.Kind)
		}
		if len(node.Interfaces) == 0 {
			add(path+".interfaces", "node has no interfaces")
		}

		seen := make(map[InterfaceID]bool, len(node.Interfaces))
		for _, ifaceID := range node.Interfaces {
			if seen[ifaceID] {
				add(path+".interfaces", "interface %q listed more than once", ifaceID)
			}
			seen[ifaceID] = true

			iface, ok := t.interfaces[ifaceID]
			if !ok {
				add(path+".interfaces", "references unknown interface %q", ifaceID)
				continue
			}
			if iface.Node != nodeID {
				add(path+".interfaces", "interface %q belongs to node %q", ifaceID, iface.Node)
			}
		}

		for idx, route := range node.Routes {
			routePath := fmt.Sprintf("%s.routes[%d]", path, idx)
			if !route.Network.IsValid() {
				add(routePath+".network", "invalid network prefix")
			}
			if route.Interface == "" {
				add(routePath+".interface", "route must name an egress interface")
			} else if iface, ok := t.interfaces[InterfaceID(route.Interface)]; !ok {
				add(routePath+".interface", "references unknown interface %q", route.Interface)
			} else if iface.Node != nodeID {
				add(routePath+".interface", "interface %q belongs to node %q", route.Interface, iface.Node)
			}
			if route.Metric < 0 {
				add(routePath+".metric", "metric must not be negative")
			}
		}
	}

	for _, ifaceID := range t.SortedInterfaceIDs() {
		iface := t.interfaces[ifaceID]
		path := fmt.Sprintf("interfaces[%s]", ifaceID)

		if !iface.Address.IsValid() {
			add(path+".address", "invalid IPv4 prefix")
			continue
		}
		if _, ok := t.nodes[iface.Node]; !ok {
			add(path+".node", "references unknown node %q", iface.Node)
		}
		if iface.Address.Addr().IsUnspecified() {
			add(path+".address", "interface address must not be 0.0.0.0")
		}
		ip := iface.IP().String()
		if other, exists := seenIPs[ip]; exists {
			add(path+".address", "duplicate address %s already used by %q", ip, other)
		} else {
			seenIPs[ip] = ifaceID
		}
	}

	for _, linkID := range t.LinkIDs() {
		link := t.links[linkID]
		path := fmt.Sprintf("links[%s]", linkID)

		a, okA := t.interfaces[link.A]
		if !okA {
			add(path+".a", "references unknown interface %q", link.A)
		}
		b, okB := t.interfaces[link.B]
		if !okB {
			add(path+".b", "references unknown interface %q", link.B)
		}
		if okA && okB && a.Node == b.Node {
			add(path, "connects node %q to itself", a.Node)
		}
		if link.Delay < 0 {
			add(path+".delay", "delay must not be negative")
		}
		if link.QueueLimit < 0 {
			add(path+".queueLimit", "queue limit must not be negative")
		}
	}

	if len(problems) == 0 {
		return nil
	}
	return &ValidationError{Problems: problems}
}
