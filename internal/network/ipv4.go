// Package network contains IPv4-like addressing, prefix matching, and static
// route tables. It is deliberately independent of topology and simulation so
// that routing rules can be tested in isolation.
package network

import (
	"fmt"
	"net/netip"
)

// IPv4 is a validated IPv4 address. The zero value is invalid.
type IPv4 struct {
	addr netip.Addr
}

// IPv4Zero is the invalid zero address.
var IPv4Zero IPv4

// ParseIPv4 parses a dotted-quad IPv4 address. IPv6 addresses are rejected.
func ParseIPv4(s string) (IPv4, error) {
	a, err := netip.ParseAddr(s)
	if err != nil {
		return IPv4{}, fmt.Errorf("network: invalid IPv4 address %q: %w", s, err)
	}
	if !a.Is4() {
		return IPv4{}, fmt.Errorf("network: %q is not an IPv4 address", s)
	}
	return IPv4{addr: a}, nil
}

// MustParseIPv4 parses s and panics on failure. It is intended for tests and
// package-level constants.
func MustParseIPv4(s string) IPv4 {
	a, err := ParseIPv4(s)
	if err != nil {
		panic(err)
	}
	return a
}

// IsValid reports whether the address is set.
func (a IPv4) IsValid() bool { return a.addr.IsValid() }

// IsUnspecified reports whether the address is 0.0.0.0.
func (a IPv4) IsUnspecified() bool { return a.addr.IsUnspecified() }

// Addr returns the underlying standard-library address.
func (a IPv4) Addr() netip.Addr { return a.addr }

// String returns the dotted-quad form, or the empty string when invalid.
func (a IPv4) String() string {
	if !a.IsValid() {
		return ""
	}
	return a.addr.String()
}

// Equal reports whether two addresses are the same.
func (a IPv4) Equal(b IPv4) bool { return a.addr == b.addr }

// Less orders addresses for deterministic output.
func (a IPv4) Less(b IPv4) bool { return a.addr.Less(b.addr) }
