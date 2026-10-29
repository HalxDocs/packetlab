package topology

import (
	"fmt"
	"sort"

	"github.com/HalxDocs/packetlab/internal/network"
)

// Topology is an indexed collection of nodes, interfaces, and links.
type Topology struct {
	nodes      map[NodeID]*Node
	interfaces map[InterfaceID]*Interface
	links      map[LinkID]*Link
	linkOf     map[InterfaceID]LinkID
}

// New returns an empty topology.
func New() *Topology {
	return &Topology{
		nodes:      make(map[NodeID]*Node),
		interfaces: make(map[InterfaceID]*Interface),
		links:      make(map[LinkID]*Link),
		linkOf:     make(map[InterfaceID]LinkID),
	}
}

// AddNode adds a node. The ID must be non-empty and unique.
func (t *Topology) AddNode(n *Node) error {
	if n == nil {
		return fmt.Errorf("topology: cannot add a nil node")
	}
	if n.ID == "" {
		return fmt.Errorf("topology: node ID must not be empty")
	}
	if _, exists := t.nodes[n.ID]; exists {
		return fmt.Errorf("topology: duplicate node ID %q", n.ID)
	}
	t.nodes[n.ID] = n
	return nil
}

// AddInterface adds an interface to an existing node.
func (t *Topology) AddInterface(i *Interface) error {
	if i == nil {
		return fmt.Errorf("topology: cannot add a nil interface")
	}
	if i.ID == "" {
		return fmt.Errorf("topology: interface ID must not be empty")
	}
	if _, exists := t.interfaces[i.ID]; exists {
		return fmt.Errorf("topology: duplicate interface ID %q", i.ID)
	}
	node, ok := t.nodes[i.Node]
	if !ok {
		return fmt.Errorf("topology: interface %q references unknown node %q", i.ID, i.Node)
	}
	t.interfaces[i.ID] = i
	if !node.HasInterface(i.ID) {
		node.Interfaces = append(node.Interfaces, i.ID)
	}
	return nil
}

// AddLink connects two existing interfaces. Each interface may have at most one
// link, and a link may not connect a node to itself.
func (t *Topology) AddLink(l *Link) error {
	if l == nil {
		return fmt.Errorf("topology: cannot add a nil link")
	}
	if l.ID == "" {
		return fmt.Errorf("topology: link ID must not be empty")
	}
	if _, exists := t.links[l.ID]; exists {
		return fmt.Errorf("topology: duplicate link ID %q", l.ID)
	}
	a, ok := t.interfaces[l.A]
	if !ok {
		return fmt.Errorf("topology: link %q references unknown interface %q", l.ID, l.A)
	}
	b, ok := t.interfaces[l.B]
	if !ok {
		return fmt.Errorf("topology: link %q references unknown interface %q", l.ID, l.B)
	}
	if a.Node == b.Node {
		return fmt.Errorf("topology: link %q connects node %q to itself", l.ID, a.Node)
	}
	if existing, ok := t.linkOf[l.A]; ok {
		return fmt.Errorf("topology: interface %q already attached to link %q", l.A, existing)
	}
	if existing, ok := t.linkOf[l.B]; ok {
		return fmt.Errorf("topology: interface %q already attached to link %q", l.B, existing)
	}
	t.links[l.ID] = l
	t.linkOf[l.A] = l.ID
	t.linkOf[l.B] = l.ID
	return nil
}

// Node returns a node by ID.
func (t *Topology) Node(id NodeID) (*Node, bool) {
	n, ok := t.nodes[id]
	return n, ok
}

// Interface returns an interface by ID.
func (t *Topology) Interface(id InterfaceID) (*Interface, bool) {
	i, ok := t.interfaces[id]
	return i, ok
}

// Link returns a link by ID.
func (t *Topology) Link(id LinkID) (*Link, bool) {
	l, ok := t.links[id]
	return l, ok
}

// LinkForInterface returns the link attached to an interface.
func (t *Topology) LinkForInterface(id InterfaceID) (*Link, bool) {
	linkID, ok := t.linkOf[id]
	if !ok {
		return nil, false
	}
	return t.Link(linkID)
}

// NodeForInterface returns the node that owns an interface.
func (t *Topology) NodeForInterface(id InterfaceID) (*Node, bool) {
	iface, ok := t.interfaces[id]
	if !ok {
		return nil, false
	}
	return t.Node(iface.Node)
}

// NodeByAddress returns the node and interface configured with ip. Results are
// deterministic because interfaces are scanned in sorted ID order.
func (t *Topology) NodeByAddress(ip network.IPv4) (*Node, *Interface, bool) {
	for _, id := range t.SortedInterfaceIDs() {
		iface := t.interfaces[id]
		if iface.Owns(ip) {
			node, ok := t.Node(iface.Node)
			if !ok {
				continue
			}
			return node, iface, true
		}
	}
	return nil, nil, false
}

// NodeIDs returns node IDs in sorted order.
func (t *Topology) NodeIDs() []NodeID {
	ids := make([]NodeID, 0, len(t.nodes))
	for id := range t.nodes {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// Nodes returns nodes in sorted ID order.
func (t *Topology) Nodes() []*Node {
	ids := t.NodeIDs()
	nodes := make([]*Node, 0, len(ids))
	for _, id := range ids {
		nodes = append(nodes, t.nodes[id])
	}
	return nodes
}

// SortedInterfaceIDs returns interface IDs in sorted order.
func (t *Topology) SortedInterfaceIDs() []InterfaceID {
	ids := make([]InterfaceID, 0, len(t.interfaces))
	for id := range t.interfaces {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// Interfaces returns interfaces in sorted ID order.
func (t *Topology) Interfaces() []*Interface {
	ids := t.SortedInterfaceIDs()
	ifaces := make([]*Interface, 0, len(ids))
	for _, id := range ids {
		ifaces = append(ifaces, t.interfaces[id])
	}
	return ifaces
}

// LinkIDs returns link IDs in sorted order.
func (t *Topology) LinkIDs() []LinkID {
	ids := make([]LinkID, 0, len(t.links))
	for id := range t.links {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// Links returns links in sorted ID order.
func (t *Topology) Links() []*Link {
	ids := t.LinkIDs()
	links := make([]*Link, 0, len(ids))
	for _, id := range ids {
		links = append(links, t.links[id])
	}
	return links
}

// NodeCount returns the number of nodes.
func (t *Topology) NodeCount() int { return len(t.nodes) }

// LinkCount returns the number of links.
func (t *Topology) LinkCount() int { return len(t.links) }
