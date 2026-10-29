package network

import "testing"

func route(network, nextHop, iface string, metric int) Route {
	return Route{
		Network:   MustParsePrefix(network),
		NextHop:   MustParseIPv4(nextHop),
		Interface: iface,
		Metric:    metric,
	}
}

func TestLookupLongestPrefixWins(t *testing.T) {
	table := NewTable([]Route{
		route("0.0.0.0/0", "10.0.1.1", "eth0", 0),
		route("10.0.2.0/24", "10.0.1.2", "eth1", 0),
		route("10.0.2.20/32", "10.0.1.3", "eth2", 0),
	})

	got, err := table.Lookup(MustParseIPv4("10.0.2.20"))
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got.Network.String() != "10.0.2.20/32" {
		t.Fatalf("selected %s, want the /32 route", got.Network)
	}

	got, err = table.Lookup(MustParseIPv4("10.0.2.21"))
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got.Network.String() != "10.0.2.0/24" {
		t.Fatalf("selected %s, want the /24 route", got.Network)
	}

	got, err = table.Lookup(MustParseIPv4("8.8.8.8"))
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if !got.Network.IsDefault() {
		t.Fatalf("selected %s, want the default route", got.Network)
	}
}

func TestLookupNoRoute(t *testing.T) {
	table := NewTable([]Route{route("10.0.1.0/24", "0.0.0.0", "eth0", 0)})
	if _, err := table.Lookup(MustParseIPv4("10.0.2.20")); err != ErrNoRoute {
		t.Fatalf("err = %v, want ErrNoRoute", err)
	}
}

func TestLookupInvalidDestination(t *testing.T) {
	table := NewTable([]Route{route("0.0.0.0/0", "10.0.1.1", "eth0", 0)})
	if _, err := table.Lookup(IPv4{}); err != ErrNoRoute {
		t.Fatalf("err = %v, want ErrNoRoute", err)
	}
}

func TestLookupTieBreaksByMetric(t *testing.T) {
	table := NewTable([]Route{
		route("10.0.0.0/8", "10.0.1.2", "eth1", 20),
		route("10.0.0.0/8", "10.0.1.1", "eth0", 10),
	})
	got, err := table.Lookup(MustParseIPv4("10.1.2.3"))
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got.Metric != 10 {
		t.Fatalf("metric = %d, want the lower metric 10", got.Metric)
	}
}

func TestLookupIsInsertionOrderIndependent(t *testing.T) {
	a := route("10.0.0.0/8", "10.0.1.1", "eth0", 10)
	b := route("10.0.0.0/8", "10.0.1.2", "eth1", 10)

	first := NewTable([]Route{a, b}).Routes()
	second := NewTable([]Route{b, a}).Routes()
	if len(first) != len(second) {
		t.Fatalf("route counts differ")
	}
	for i := range first {
		if first[i].NextHop.String() != second[i].NextHop.String() {
			t.Fatalf("order differs at %d: %s vs %s", i, first[i].NextHop, second[i].NextHop)
		}
	}
}

func TestDirectlyConnected(t *testing.T) {
	connected := Route{Network: MustParsePrefix("10.0.1.0/24"), NextHop: IPv4{}}
	if !connected.DirectlyConnected() {
		t.Fatal("zero next hop should mean directly connected")
	}
	viaRouter := route("0.0.0.0/0", "10.0.1.1", "eth0", 0)
	if viaRouter.DirectlyConnected() {
		t.Fatal("route with a next hop is not directly connected")
	}
}
