package network

import (
	"fmt"
	"net/netip"
)

// Prefix is an IPv4 network prefix such as 10.0.1.0/24. The address may carry
// host bits, as it does for an interface address like 10.0.1.10/24; Masked
// returns the canonical network form.
type Prefix struct {
	prefix netip.Prefix
}

// ParsePrefix parses an IPv4 prefix in CIDR notation.
func ParsePrefix(s string) (Prefix, error) {
	p, err := netip.ParsePrefix(s)
	if err != nil {
		return Prefix{}, fmt.Errorf("network: invalid prefix %q: %w", s, err)
	}
	if !p.Addr().Is4() {
		return Prefix{}, fmt.Errorf("network: %q is not an IPv4 prefix", s)
	}
	return Prefix{prefix: p}, nil
}

// MustParsePrefix parses s and panics on failure.
func MustParsePrefix(s string) Prefix {
	p, err := ParsePrefix(s)
	if err != nil {
		panic(err)
	}
	return p
}

// IsValid reports whether the prefix is set.
func (p Prefix) IsValid() bool { return p.prefix.IsValid() }

// Bits returns the prefix length.
func (p Prefix) Bits() int { return p.prefix.Bits() }

// Addr returns the address as written, including any host bits.
func (p Prefix) Addr() IPv4 { return IPv4{addr: p.prefix.Addr()} }

// Masked returns the canonical network address with host bits cleared.
func (p Prefix) Masked() Prefix { return Prefix{prefix: p.prefix.Masked()} }

// Contains reports whether the address falls inside the prefix.
func (p Prefix) Contains(a IPv4) bool {
	if !p.IsValid() || !a.IsValid() {
		return false
	}
	return p.prefix.Contains(a.addr)
}

// IsDefault reports whether this is the default route prefix 0.0.0.0/0.
func (p Prefix) IsDefault() bool {
	return p.IsValid() && p.prefix.Bits() == 0
}

// String returns CIDR notation.
func (p Prefix) String() string {
	if !p.IsValid() {
		return ""
	}
	return p.prefix.String()
}
