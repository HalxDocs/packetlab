package topology

import "github.com/HalxDocs/packetlab/internal/sim"

// Link is a point-to-point connection between two interfaces.
type Link struct {
	ID           LinkID
	A            InterfaceID
	B            InterfaceID
	Delay        sim.SimTime
	BandwidthBps uint64
	QueueLimit   int
	Up           bool
}

// Peer returns the interface at the other end of the link.
func (l *Link) Peer(iface InterfaceID) (InterfaceID, bool) {
	switch iface {
	case l.A:
		return l.B, true
	case l.B:
		return l.A, true
	default:
		return "", false
	}
}

// Connects reports whether the link attaches to the interface.
func (l *Link) Connects(iface InterfaceID) bool {
	return l.A == iface || l.B == iface
}
