package eventlog

import (
	"testing"

	"github.com/HalxDocs/packetlab/internal/sim"
)

func TestDescribe(t *testing.T) {
	tests := []struct {
		name  string
		event sim.Event
		want  string
	}{
		{
			name:  "created echo request",
			event: sim.NewEvent(0, sim.PacketCreated).WithDetail("icmp", "echo_request"),
			want:  "created ICMP echo request",
		},
		{
			name:  "created echo reply",
			event: sim.NewEvent(0, sim.PacketCreated).WithDetail("icmp", "echo_reply"),
			want:  "created ICMP echo reply",
		},
		{
			name:  "created packet",
			event: sim.NewEvent(0, sim.PacketCreated),
			want:  "created packet",
		},
		{
			name: "route via next hop",
			event: sim.NewEvent(0, sim.RouteSelected).
				WithDetail("network", "0.0.0.0/0").
				WithDetail("nextHop", "10.0.1.1").
				WithDetail("interface", "host-a:eth0"),
			want: "selected 0.0.0.0/0 via 10.0.1.1 on host-a:eth0",
		},
		{
			name: "direct route",
			event: sim.NewEvent(0, sim.RouteSelected).
				WithDetail("network", "10.0.2.0/24").
				WithDetail("nextHop", "0.0.0.0").
				WithDetail("interface", "router-r1:eth1"),
			want: "selected 10.0.2.0/24 directly on router-r1:eth1",
		},
		{
			name:  "transmit",
			event: sim.NewEvent(0, sim.LinkTransmit).WithLink("link-a"),
			want:  "transmitted on link-a",
		},
		{
			name:  "delivered on link",
			event: sim.NewEvent(0, sim.LinkDelivered).WithDetail("ingress", "router-r1:eth0"),
			want:  "arrived on router-r1:eth0",
		},
		{
			name:  "received",
			event: sim.NewEvent(0, sim.PacketReceived).WithDetail("ingress", "router-r1:eth0"),
			want:  "received on router-r1:eth0",
		},
		{
			name: "forwarded",
			event: sim.NewEvent(0, sim.PacketForwarded).
				WithDetail("ingress", "router-r1:eth0").
				WithDetail("egress", "router-r1:eth1"),
			want: "forwarded router-r1:eth0 -> router-r1:eth1",
		},
		{
			name:  "delivered to host",
			event: sim.NewEvent(0, sim.PacketDelivered).WithNode("host-b"),
			want:  "delivered to host-b",
		},
		{
			name:  "dropped no route",
			event: sim.NewEvent(0, sim.PacketDropped).WithReason(sim.ReasonNoRoute),
			want:  "dropped: no route to destination",
		},
		{
			name:  "ping reply",
			event: sim.NewEvent(0, sim.PingReply).WithDetail("rttMs", "2.000"),
			want:  "echo reply received (rtt 2.000ms)",
		},
		{
			name:  "run started",
			event: sim.NewEvent(0, sim.RunStarted),
			want:  "simulation started",
		},
		{
			name:  "run completed",
			event: sim.NewEvent(0, sim.RunCompleted),
			want:  "simulation completed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Describe(tt.event); got != tt.want {
				t.Fatalf("Describe = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReasonText(t *testing.T) {
	cases := map[sim.ReasonCode]string{
		sim.ReasonNoRoute:            "no route to destination",
		sim.ReasonLinkDown:           "link is down",
		sim.ReasonTTLExpired:         "TTL expired",
		sim.ReasonNoInterface:        "route has no usable interface",
		sim.ReasonUnknownDestination: "destination address is not local",
		sim.ReasonNodeDown:           "node is down",
		sim.ReasonQueueFull:          "queue is full",
		sim.ReasonLinkLoss:           "link loss model dropped the packet",
	}
	for code, want := range cases {
		if got := ReasonText(code); got != want {
			t.Errorf("ReasonText(%q) = %q, want %q", code, got, want)
		}
	}
}
