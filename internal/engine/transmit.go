package engine

import (
	"fmt"

	"github.com/HalxDocs/packetlab/internal/packet"
	"github.com/HalxDocs/packetlab/internal/sim"
	"github.com/HalxDocs/packetlab/internal/topology"
)

// transmit puts a packet on the link attached to egress. It records the
// attempt and schedules the arrival after serialization plus propagation.
func (e *Engine) transmit(nodeID topology.NodeID, egress topology.InterfaceID, pkt packet.Packet) error {
	link, ok := e.topo.LinkForInterface(egress)
	if !ok {
		e.drop(nodeID, pkt, sim.ReasonNoInterface, "", map[string]string{
			"egress": string(egress),
		})
		return nil
	}
	if !link.Up {
		e.drop(nodeID, pkt, sim.ReasonLinkDown, link.ID, map[string]string{
			"egress": string(egress),
		})
		return nil
	}
	peer, ok := link.Peer(egress)
	if !ok {
		e.drop(nodeID, pkt, sim.ReasonNoInterface, link.ID, map[string]string{
			"egress": string(egress),
		})
		return nil
	}

	size := pkt.SizeBytes()
	serialization := serializationDelay(link, size)
	now := e.clock.Now()

	e.record(sim.NewEvent(now, sim.LinkTransmit).
		WithNode(string(nodeID)).
		WithLink(string(link.ID)).
		WithPacket(string(pkt.ID)).
		WithDetail("egress", string(egress)).
		WithDetail("delay", link.Delay.String()).
		WithDetail("serialization", serialization.String()).
		WithDetail("bytes", fmt.Sprintf("%d", size)))

	e.schedule(now.Add(serialization).Add(link.Delay), eventLinkArrive, linkArrival{
		Link:    link.ID,
		Ingress: peer,
		Egress:  egress,
		Packet:  pkt.Clone(),
	})
	return nil
}

// serializationDelay returns the time needed to place sizeBytes on the link.
// A zero bandwidth means unlimited.
func serializationDelay(link *topology.Link, sizeBytes int) sim.SimTime {
	if link.BandwidthBps == 0 || sizeBytes <= 0 {
		return 0
	}
	bits := uint64(sizeBytes) * 8
	ns := (bits*1_000_000_000 + link.BandwidthBps - 1) / link.BandwidthBps
	return sim.SimTime(ns)
}
