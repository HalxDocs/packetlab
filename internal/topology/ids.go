// Package topology defines nodes, interfaces, links, and structural
// validation. It deliberately knows nothing about packet forwarding.
package topology

// NodeID identifies a node in a topology.
type NodeID string

// InterfaceID identifies a network interface on a node.
type InterfaceID string

// LinkID identifies a point-to-point link.
type LinkID string

// NodeKind describes how a node handles packets.
type NodeKind string

const (
	// KindHost terminates packets. Hosts do not forward.
	KindHost NodeKind = "host"
	// KindRouter forwards packets between interfaces.
	KindRouter NodeKind = "router"
)

// Valid reports whether the kind is known.
func (k NodeKind) Valid() bool {
	return k == KindHost || k == KindRouter
}

// Forwards reports whether nodes of this kind forward packets.
func (k NodeKind) Forwards() bool { return k == KindRouter }
