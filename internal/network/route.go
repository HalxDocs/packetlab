package network

import (
	"errors"
	"sort"
)

// ErrNoRoute is returned when no route matches a destination.
var ErrNoRoute = errors.New("network: no route to destination")

// Route is one entry in a static route table.
type Route struct {
	Network   Prefix
	NextHop   IPv4
	Interface string
	Metric    int
}

// DirectlyConnected reports whether the route has no next hop, meaning the
// destination is on a directly attached network.
func (r Route) DirectlyConnected() bool {
	return !r.NextHop.IsValid() || r.NextHop.IsUnspecified()
}

// Table is a static route table. Routes are sorted once at construction so
// lookup and output never depend on insertion order.
type Table struct {
	routes []Route
}

// NewTable returns a table containing routes. Lookup uses longest-prefix
// match; ties are broken by metric, then by stable textual fields.
func NewTable(routes []Route) *Table {
	sorted := append([]Route(nil), routes...)
	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		if a.Network.Bits() != b.Network.Bits() {
			return a.Network.Bits() > b.Network.Bits()
		}
		if a.Metric != b.Metric {
			return a.Metric < b.Metric
		}
		if a.Network.String() != b.Network.String() {
			return a.Network.String() < b.Network.String()
		}
		if a.NextHop.String() != b.NextHop.String() {
			return a.NextHop.String() < b.NextHop.String()
		}
		return a.Interface < b.Interface
	})
	return &Table{routes: sorted}
}

// Lookup returns the best route for dst, or ErrNoRoute.
func (t *Table) Lookup(dst IPv4) (Route, error) {
	if !dst.IsValid() {
		return Route{}, ErrNoRoute
	}
	for _, r := range t.routes {
		if r.Network.Contains(dst) {
			return r, nil
		}
	}
	return Route{}, ErrNoRoute
}

// Routes returns a copy of the sorted table.
func (t *Table) Routes() []Route {
	return append([]Route(nil), t.routes...)
}

// Len returns the number of routes.
func (t *Table) Len() int { return len(t.routes) }
