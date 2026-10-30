package engine

import (
	"fmt"

	"github.com/HalxDocs/packetlab/internal/network"
	"github.com/HalxDocs/packetlab/internal/packet"
	"github.com/HalxDocs/packetlab/internal/protocols/icmp"
	"github.com/HalxDocs/packetlab/internal/sim"
	"github.com/HalxDocs/packetlab/internal/topology"
)

// pingState correlates an echo request with its reply.
type pingState struct {
	requestID packet.ID
	sentAt    sim.SimTime
	from      topology.NodeID
	to        network.IPv4
}

// pingKey combines the echo identifier and sequence into one map key.
func pingKey(id, seq uint16) uint32 {
	return uint32(id)<<16 | uint32(seq)
}

// startPing creates an echo request, selects a route, and transmits it.
func (e *Engine) startPing(p scheduledPing) error {
	node, ok := e.topo.Node(p.From)
	if !ok {
		return fmt.Errorf("engine: ping source %q no longer exists", p.From)
	}

	route, routeErr := e.tables[node.ID].Lookup(p.To)
	egress := topology.InterfaceID(route.Interface)
	var src network.IPv4
	if routeErr == nil {
		if iface, ok := e.topo.Interface(egress); ok {
			src = iface.IP()
		}
	}

	now := e.clock.Now()
	pkt := packet.Packet{
		ID:        e.ids.Next(),
		Src:       src,
		Dst:       p.To,
		Protocol:  packet.ProtocolICMP,
		TTL:       p.TTL,
		Payload:   icmp.Encode(icmp.Request(p.EchoID, p.Seq, nil)),
		CreatedAt: now,
	}

	e.record(sim.NewEvent(now, sim.PacketCreated).
		WithNode(string(node.ID)).
		WithPacket(string(pkt.ID)).
		WithDetail("src", src.String()).
		WithDetail("dst", p.To.String()).
		WithDetail("protocol", "icmp").
		WithDetail("ttl", fmt.Sprintf("%d", p.TTL)).
		WithDetail("icmp", "echo_request"))

	if routeErr != nil {
		e.drop(node.ID, pkt, sim.ReasonNoRoute, "", map[string]string{"dst": p.To.String()})
		return nil
	}
	if !src.IsValid() {
		e.drop(node.ID, pkt, sim.ReasonNoInterface, "", map[string]string{"interface": route.Interface})
		return nil
	}

	e.recordRoute(node, pkt, route)
	e.pingState[pingKey(p.EchoID, p.Seq)] = &pingState{
		requestID: pkt.ID,
		sentAt:    now,
		from:      p.From,
		to:        p.To,
	}

	return e.transmit(node.ID, egress, pkt)
}

// completePing records a successful round trip.
func (e *Engine) completePing(node *topology.Node, reply packet.Packet, msg icmp.Message) {
	key := pingKey(msg.Echo.ID, msg.Echo.Seq)
	state, ok := e.pingState[key]
	if !ok {
		return
	}
	delete(e.pingState, key)

	rtt := e.clock.Now().Sub(state.sentAt)
	e.record(sim.NewEvent(e.clock.Now(), sim.PingReply).
		WithNode(string(node.ID)).
		WithPacket(string(reply.ID)).
		WithDetail("request", string(state.requestID)).
		WithDetail("reply", string(reply.ID)).
		WithDetail("seq", fmt.Sprintf("%d", msg.Echo.Seq)).
		WithDetail("rttMs", fmt.Sprintf("%.3f", rtt.Milliseconds())))
}
