package eventlog

import (
	"fmt"

	"github.com/HalxDocs/packetlab/internal/sim"
)

// Describe returns the human-readable summary for an event. Summaries are
// derived only from the event itself so they are deterministic; they are not
// part of the stable machine contract (kinds and reason codes are).
func Describe(e sim.Event) string {
	switch e.Kind {
	case sim.RunStarted:
		return "simulation started"
	case sim.RunCompleted:
		return "simulation completed"
	case sim.PacketCreated:
		switch e.Details["icmp"] {
		case "echo_request":
			return "created ICMP echo request"
		case "echo_reply":
			return "created ICMP echo reply"
		default:
			return "created packet"
		}
	case sim.RouteSelected:
		network := e.Details["network"]
		nextHop := e.Details["nextHop"]
		iface := e.Details["interface"]
		if nextHop == "" || nextHop == "0.0.0.0" {
			return fmt.Sprintf("selected %s directly on %s", network, iface)
		}
		return fmt.Sprintf("selected %s via %s on %s", network, nextHop, iface)
	case sim.LinkTransmit:
		return fmt.Sprintf("transmitted on %s", e.LinkID)
	case sim.LinkDelivered:
		return fmt.Sprintf("arrived on %s", e.Details["ingress"])
	case sim.PacketReceived:
		return fmt.Sprintf("received on %s", e.Details["ingress"])
	case sim.PacketForwarded:
		return fmt.Sprintf("forwarded %s -> %s", e.Details["ingress"], e.Details["egress"])
	case sim.PacketDelivered:
		return fmt.Sprintf("delivered to %s", e.NodeID)
	case sim.PacketDropped:
		return "dropped: " + ReasonText(e.Reason)
	case sim.PingRequested:
		return fmt.Sprintf("ping %s", e.Details["dst"])
	case sim.PingReply:
		return fmt.Sprintf("echo reply received (rtt %sms)", e.Details["rttMs"])
	case sim.LinkStateChanged:
		return fmt.Sprintf("link %s", e.Details["state"])
	default:
		return string(e.Kind)
	}
}

// ReasonText translates a stable reason code into an explanation for humans.
func ReasonText(r sim.ReasonCode) string {
	switch r {
	case sim.ReasonNoRoute:
		return "no route to destination"
	case sim.ReasonLinkDown:
		return "link is down"
	case sim.ReasonTTLExpired:
		return "TTL expired"
	case sim.ReasonNoInterface:
		return "route has no usable interface"
	case sim.ReasonUnknownDestination:
		return "destination address is not local"
	case sim.ReasonNodeDown:
		return "node is down"
	case sim.ReasonQueueFull:
		return "queue is full"
	case sim.ReasonLinkLoss:
		return "link loss model dropped the packet"
	default:
		return string(r)
	}
}
