package topology

import "github.com/HalxDocs/packetlab/internal/network"

// Interface is a node's attachment to a link, with an address and prefix.
type Interface struct {
	ID      InterfaceID
	Node    NodeID
	Address network.Prefix
}

// IP returns the interface address.
func (i *Interface) IP() network.IPv4 { return i.Address.Addr() }

// Subnet returns the canonical network prefix for the interface.
func (i *Interface) Subnet() network.Prefix { return i.Address.Masked() }

// Owns reports whether the interface is configured with exactly this address.
func (i *Interface) Owns(ip network.IPv4) bool {
	return i.Address.IsValid() && i.Address.Addr().Equal(ip)
}
