package sim

import "testing"

func TestClockStartsAtZero(t *testing.T) {
	if got := NewClock().Now(); got != 0 {
		t.Fatalf("Now = %s, want 0", got)
	}
}

func TestClockAdvances(t *testing.T) {
	c := NewClock()
	if err := c.AdvanceTo(Ms(5)); err != nil {
		t.Fatalf("AdvanceTo: %v", err)
	}
	if got := c.Now(); got != Ms(5) {
		t.Fatalf("Now = %s, want 5ms", got)
	}
}

func TestClockRejectsBackwardTime(t *testing.T) {
	c := NewClock()
	if err := c.AdvanceTo(Ms(5)); err != nil {
		t.Fatalf("AdvanceTo: %v", err)
	}
	if err := c.AdvanceTo(Ms(4)); err == nil {
		t.Fatal("expected error moving time backward")
	}
	if got := c.Now(); got != Ms(5) {
		t.Fatalf("Now = %s after rejected move, want 5ms", got)
	}
}

func TestClockAllowsSameTime(t *testing.T) {
	c := NewClock()
	if err := c.AdvanceTo(Ms(5)); err != nil {
		t.Fatalf("AdvanceTo: %v", err)
	}
	if err := c.AdvanceTo(Ms(5)); err != nil {
		t.Fatalf("AdvanceTo same time: %v", err)
	}
}

func TestClockReset(t *testing.T) {
	c := NewClock()
	if err := c.AdvanceTo(Ms(5)); err != nil {
		t.Fatalf("AdvanceTo: %v", err)
	}
	c.Reset()
	if got := c.Now(); got != 0 {
		t.Fatalf("Now after Reset = %s, want 0", got)
	}
}
