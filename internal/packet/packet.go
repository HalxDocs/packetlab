package packet

import (
	"github.com/HalxDocs/packetlab/internal/network"
	"github.com/HalxDocs/packetlab/internal/sim"
)

// Protocol is an IP protocol number.
type Protocol uint8

// Protocol numbers used by PacketLab.
const (
	ProtocolICMP Protocol = 1
	ProtocolTCP  Protocol = 6
	ProtocolUDP  Protocol = 17
)

// String returns the protocol name used in traces.
func (p Protocol) String() string {
	switch p {
	case ProtocolICMP:
		return "icmp"
	case ProtocolTCP:
		return "tcp"
	case ProtocolUDP:
		return "udp"
	default:
		return "unknown"
	}
}

// Header sizes used to estimate serialization delay. PacketLab does not model
// Ethernet framing, so this is an explicit simplification.
const (
	IPv4HeaderBytes = 20
	ICMPHeaderBytes = 8
)

// Packet is a simplified IPv4 packet. It carries only the fields the current
// protocol slice needs.
type Packet struct {
	ID        ID
	Src       network.IPv4
	Dst       network.IPv4
	Protocol  Protocol
	TTL       uint8
	Payload   []byte
	CreatedAt sim.SimTime
}

// Clone returns a deep copy of the packet.
func (p Packet) Clone() Packet {
	cp := p
	cp.Payload = append([]byte(nil), p.Payload...)
	return cp
}

// DecrementTTL returns a copy with the TTL reduced by one. ok is false when the
// packet has no hops left and must be dropped.
func (p Packet) DecrementTTL() (Packet, bool) {
	if p.TTL <= 1 {
		return p, false
	}
	p.TTL--
	return p, true
}

// SizeBytes estimates the on-wire size used for serialization delay.
func (p Packet) SizeBytes() int {
	return IPv4HeaderBytes + ICMPHeaderBytes + len(p.Payload)
}
