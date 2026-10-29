package network

import "testing"

func TestParseIPv4(t *testing.T) {
	valid := []string{"10.0.1.10", "0.0.0.0", "255.255.255.255", "192.168.0.1"}
	for _, s := range valid {
		a, err := ParseIPv4(s)
		if err != nil {
			t.Errorf("ParseIPv4(%q): %v", s, err)
			continue
		}
		if got := a.String(); got != s {
			t.Errorf("ParseIPv4(%q).String() = %q", s, got)
		}
		if !a.IsValid() {
			t.Errorf("ParseIPv4(%q).IsValid() = false", s)
		}
	}

	invalid := []string{"", "10.0.1", "10.0.1.256", "::1", "2001:db8::1", "host"}
	for _, s := range invalid {
		if _, err := ParseIPv4(s); err == nil {
			t.Errorf("ParseIPv4(%q) succeeded, want error", s)
		}
	}
}

func TestIPv4ZeroValue(t *testing.T) {
	var a IPv4
	if a.IsValid() {
		t.Fatal("zero value should be invalid")
	}
	if got := a.String(); got != "" {
		t.Fatalf("zero value String() = %q, want empty", got)
	}
}

func TestIPv4Unspecified(t *testing.T) {
	if !MustParseIPv4("0.0.0.0").IsUnspecified() {
		t.Fatal("0.0.0.0 should be unspecified")
	}
	if MustParseIPv4("10.0.0.1").IsUnspecified() {
		t.Fatal("10.0.0.1 should not be unspecified")
	}
}

func TestIPv4Ordering(t *testing.T) {
	a := MustParseIPv4("10.0.0.1")
	b := MustParseIPv4("10.0.0.2")
	if !a.Less(b) || b.Less(a) {
		t.Fatal("expected 10.0.0.1 < 10.0.0.2")
	}
	if !a.Equal(MustParseIPv4("10.0.0.1")) {
		t.Fatal("expected equal addresses")
	}
}
