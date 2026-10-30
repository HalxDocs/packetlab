package engine

import (
	"fmt"

	"github.com/HalxDocs/packetlab/internal/network"
	"github.com/HalxDocs/packetlab/internal/packet"
	"github.com/HalxDocs/packetlab/internal/protocols/icmp"
	"github.com/HalxDocs/packetlab/internal/sim"
	"github.com/HalxDocs/packetlab/internal/topology"
)

// arrive handles a completed link transmission.
func (e *Engine) arrive(a linkArrival) error {
	link, ok := e.topo.Link(a.Link)
	if !ok {
		return fmt.Errorf("engine: arrival on unknown link %q", a.Link)
	}
	node, ok := e.topo.NodeForInterface(a.Ingress)
	if !ok {
		return fmt.Errorf("engine: arrival on unknown interface %q", a.Ingress)
	}

	e.record(sim.NewEvent(e.clock.Now(), sim.LinkDelivered).
		WithNode(string(node.ID)).
		WithLink(string(link.ID)).
		WithPacket(string(a.Packet.ID)).
		WithDetail("ingress", string(a.Ingress)))

	if node.Kind.Forwards() {
		return e.receiveRouter(node, a.Ingress, a.Packet)
	}
	return e.receiveHost(node, a.Ingress, a.Packet)
}

// receiveRouter applies TTL handling and forwards the packet.
func (e *Engine) receiveRouter(node *topology.Node, ingress topology.InterfaceID, pkt packet.Packet) error {
	e.record(sim.NewEvent(e.clock.Now(), sim.PacketReceived).
		WithNode(string(node.ID)).
		WithPacket(string(pkt.ID)).
		WithDetail("ingress", string(ingress)).
		WithDetail("ttl", fmt.Sprintf("%d", pkt.TTL)))

	if node.Down {
		e.drop(node.ID, pkt, sim.ReasonNodeDown, "", map[string]string{"ingress": string(ingress)})
		return nil
	}

	next, ok := pkt.DecrementTTL()
	if !ok {
		e.drop(node.ID, pkt, sim.ReasonTTLExpired, "", map[string]string{
			"ingress": string(ingress),
			"ttl":     fmt.Sprintf("%d", pkt.TTL),
		})
		return nil
	}

	route, err := e.tables[node.ID].Lookup(next.Dst)
	if err != nil {
		e.drop(node.ID, next, sim.ReasonNoRoute, "", map[string]string{
			"ingress": string(ingress),
			"dst":     next.Dst.String(),
		})
		return nil
	}
	e.recordRoute(node, next, route)

	e.record(sim.NewEvent(e.clock.Now(), sim.PacketForwarded).
		WithNode(string(node.ID)).
		WithPacket(string(next.ID)).
		WithDetail("ingress", string(ingress)).
		WithDetail("egress", route.Interface).
		WithDetail("ttl", fmt.Sprintf("%d", next.TTL)))

	return e.transmit(node.ID, topology.InterfaceID(route.Interface), next)
}

// receiveHost accepts packets addressed to the node and handles ICMP echo.
func (e *Engine) receiveHost(node *topology.Node, ingress topology.InterfaceID, pkt packet.Packet) error {
	e.record(sim.NewEvent(e.clock.Now(), sim.PacketReceived).
		WithNode(string(node.ID)).
		WithPacket(string(pkt.ID)).
		WithDetail("ingress", string(ingress)).
		WithDetail("ttl", fmt.Sprintf("%d", pkt.TTL)))

	if node.Down {
		e.drop(node.ID, pkt, sim.ReasonNodeDown, "", map[string]string{"ingress": string(ingress)})
		return nil
	}

	owner, iface, ok := e.topo.NodeByAddress(pkt.Dst)
	if !ok || owner.ID != node.ID {
		e.drop(node.ID, pkt, sim.ReasonUnknownDestination, "", map[string]string{
			"ingress": string(ingress),
			"dst":     pkt.Dst.String(),
		})
		return nil
	}

	e.record(sim.NewEvent(e.clock.Now(), sim.PacketDelivered).
		WithNode(string(node.ID)).
		WithPacket(string(pkt.ID)).
		WithDetail("ingress", string(ingress)).
		WithDetail("dst", pkt.Dst.String()).
		WithDetail("protocol", pkt.Protocol.String()))
	e.delivered++

	if pkt.Protocol != packet.ProtocolICMP {
		return nil
	}
	msg, err := icmp.Decode(pkt.Payload)
	if err != nil {
		return nil
	}
	switch {
	case msg.IsRequest():
		return e.replyToEcho(node, iface, pkt, msg)
	case msg.IsReply():
		e.completePing(node, pkt, msg)
	}
	return nil
}

// replyToEcho creates and sends an echo reply for a request.
func (e *Engine) replyToEcho(node *topology.Node, iface *topology.Interface, request packet.Packet, msg icmp.Message) error {
	now := e.clock.Now()
	reply := packet.Packet{
		ID:        e.ids.Next(),
		Src:       request.Dst,
		Dst:       request.Src,
		Protocol:  packet.ProtocolICMP,
		TTL:       DefaultTTL,
		Payload:   icmp.Encode(icmp.ReplyTo(msg)),
		CreatedAt: now,
	}

	e.record(sim.NewEvent(now, sim.PacketCreated).
		WithNode(string(node.ID)).
		WithPacket(string(reply.ID)).
		WithDetail("src", reply.Src.String()).
		WithDetail("dst", reply.Dst.String()).
		WithDetail("protocol", reply.Protocol.String()).
		WithDetail("ttl", fmt.Sprintf("%d", reply.TTL)).
		WithDetail("icmp", "echo_reply").
		WithDetail("replyTo", string(request.ID)))

	route, err := e.tables[node.ID].Lookup(reply.Dst)
	if err != nil {
		e.drop(node.ID, reply, sim.ReasonNoRoute, "", map[string]string{"dst": reply.Dst.String()})
		return nil
	}
	e.recordRoute(node, reply, route)

	return e.transmit(node.ID, topology.InterfaceID(route.Interface), reply)
}

// recordRoute logs a successful route selection.
func (e *Engine) recordRoute(node *topology.Node, pkt packet.Packet, route network.Route) {
	nextHop := route.NextHop.String()
	if route.DirectlyConnected() {
		nextHop = "0.0.0.0"
	}
	e.record(sim.NewEvent(e.clock.Now(), sim.RouteSelected).
		WithNode(string(node.ID)).
		WithPacket(string(pkt.ID)).
		WithDetail("network", route.Network.String()).
		WithDetail("nextHop", nextHop).
		WithDetail("interface", route.Interface).
		WithDetail("metric", fmt.Sprintf("%d", route.Metric)))
}

// drop records a packet failure with a stable reason code.
func (e *Engine) drop(nodeID topology.NodeID, pkt packet.Packet, reason sim.ReasonCode, linkID topology.LinkID, details map[string]string) {
	e.dropped++
	ev := sim.NewEvent(e.clock.Now(), sim.PacketDropped).
		WithNode(string(nodeID)).
		WithPacket(string(pkt.ID)).
		WithReason(reason)
	if linkID != "" {
		ev = ev.WithLink(string(linkID))
	}
	for k, v := range details {
		ev = ev.WithDetail(k, v)
	}
	e.record(ev)
}
