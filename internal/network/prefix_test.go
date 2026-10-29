package network

import "testing"

func TestParsePrefix(t *testing.T) {
	p, err := ParsePrefix("10.0.1.10/24")
	if err != nil {
		t.Fatalf("ParsePrefix: %v", err)
	}
	if got := p.Addr().String(); got != "10.0.1.10" {
		t.Errorf("Addr = %q, want 10.0.1.10", got)
	}
	if got := p.Bits(); got != 24 {
		t.Errorf("Bits = %d, want 24", got)
	}
	if got := p.Masked().String(); got != "10.0.1.0/24" {
		t.Errorf("Masked = %q, want 10.0.1.0/24", got)
	}
}

func TestParsePrefixRejectsInvalid(t *testing.T) {
	for _, s := range []string{"", "10.0.1.1", "10.0.1.1/33", "::1/128", "10.0.1.1/-1"} {
		if _, err := ParsePrefix(s); err == nil {
			t.Errorf("ParsePrefix(%q) succeeded, want error", s)
		}
	}
}

func TestPrefixContains(t *testing.T) {
	p := MustParsePrefix("10.0.1.0/24")
	if !p.Contains(MustParseIPv4("10.0.1.10")) {
		t.Error("10.0.1.10 should be in 10.0.1.0/24")
	}
	if !p.Contains(MustParseIPv4("10.0.1.0")) {
		t.Error("network address should be in prefix")
	}
	if !p.Contains(MustParseIPv4("10.0.1.255")) {
		t.Error("broadcast address should be in prefix")
	}
	if p.Contains(MustParseIPv4("10.0.2.1")) {
		t.Error("10.0.2.1 should not be in 10.0.1.0/24")
	}
}

func TestPrefixHostBitsDoNotChangeMatching(t *testing.T) {
	p := MustParsePrefix("10.0.1.10/24")
	if !p.Contains(MustParseIPv4("10.0.1.99")) {
		t.Error("unmasked interface prefix should still match the whole subnet")
	}
	if p.Contains(MustParseIPv4("10.0.2.10")) {
		t.Error("prefix should not match another subnet")
	}
}

func TestDefaultPrefix(t *testing.T) {
	if !MustParsePrefix("0.0.0.0/0").IsDefault() {
		t.Error("0.0.0.0/0 should be the default prefix")
	}
	if MustParsePrefix("10.0.0.0/8").IsDefault() {
		t.Error("10.0.0.0/8 should not be the default prefix")
	}
}

func TestInvalidPrefixMatching(t *testing.T) {
	var p Prefix
	if p.Contains(MustParseIPv4("10.0.0.1")) {
		t.Fatal("invalid prefix should match nothing")
	}
}
