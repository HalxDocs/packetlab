package topology

import "github.com/HalxDocs/packetlab/internal/network"

// Node is a host or router.
type Node struct {
	ID         NodeID
	Name       string
	Kind       NodeKind
	Interfaces []InterfaceID
	Routes     []network.Route
	Down       bool
}

// DisplayName returns the human-readable name, falling back to the ID.
func (n *Node) DisplayName() string {
	if n.Name != "" {
		return n.Name
	}
	return string(n.ID)
}

// HasInterface reports whether the node lists the interface.
func (n *Node) HasInterface(id InterfaceID) bool {
	for _, have := range n.Interfaces {
		if have == id {
			return true
		}
	}
	return false
}
