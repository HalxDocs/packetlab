package engine

import (
	"github.com/HalxDocs/packetlab/internal/packet"
	"github.com/HalxDocs/packetlab/internal/sim"
	"github.com/HalxDocs/packetlab/internal/topology"
)

// Internal event kinds. These never appear in the log; they are causes that
// handlers turn into records.
const (
	eventSendPing   sim.EventKind = "engine.send_ping"
	eventLinkArrive sim.EventKind = "engine.link_arrive"
)

// linkArrival is the payload of a scheduled link delivery.
type linkArrival struct {
	Link    topology.LinkID
	Ingress topology.InterfaceID
	Egress  topology.InterfaceID
	Packet  packet.Packet
}
