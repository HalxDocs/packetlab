package sim

import "testing"

func TestSimTimeUnits(t *testing.T) {
	if got := Ms(2); got != 2*Millisecond {
		t.Fatalf("Ms(2) = %d, want %d", got, 2*Millisecond)
	}
	if got := Us(500); got != 500*Microsecond {
		t.Fatalf("Us(500) = %d, want %d", got, 500*Microsecond)
	}
	if got := Millisecond; got != 1_000_000 {
		t.Fatalf("Millisecond = %d, want 1000000", got)
	}
}

func TestSimTimeString(t *testing.T) {
	tests := []struct {
		in   SimTime
		want string
	}{
		{0, "0ms"},
		{500 * Microsecond, "500µs"},
		{1500 * Microsecond, "1.5ms"},
		{2 * Millisecond, "2ms"},
		{2500 * Millisecond, "2.5s"},
		{-2 * Millisecond, "-2ms"},
	}
	for _, tt := range tests {
		if got := tt.in.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSimTimeTraceString(t *testing.T) {
	tests := []struct {
		in   SimTime
		want string
	}{
		{0, "0.000ms"},
		{Ms(1), "1.000ms"},
		{1500 * Microsecond, "1.500ms"},
		{2 * Second, "2000.000ms"},
	}
	for _, tt := range tests {
		if got := tt.in.TraceString(); got != tt.want {
			t.Errorf("%d.TraceString() = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSimTimeArithmetic(t *testing.T) {
	base := Ms(10)
	if got := base.Add(Ms(5)); got != Ms(15) {
		t.Errorf("Add = %s, want 15ms", got)
	}
	if got := base.Sub(Ms(15)); got != Ms(-5) {
		t.Errorf("Sub = %s, want -5ms", got)
	}
	if !Ms(1).Before(Ms(2)) {
		t.Error("expected 1ms before 2ms")
	}
	if Ms(2).Before(Ms(2)) {
		t.Error("expected 2ms not before itself")
	}
}
