package packet

import "testing"

func TestAllocatorProducesSequentialIDs(t *testing.T) {
	a := NewAllocator()
	for i, want := range []ID{"p-001", "p-002", "p-003"} {
		if got := a.Next(); got != want {
			t.Fatalf("allocation %d = %q, want %q", i, got, want)
		}
	}
	if got := a.Allocated(); got != 3 {
		t.Fatalf("Allocated = %d, want 3", got)
	}
}

func TestAllocatorReset(t *testing.T) {
	a := NewAllocator()
	a.Next()
	a.Next()
	a.Reset()
	if got := a.Next(); got != "p-001" {
		t.Fatalf("after Reset = %q, want p-001", got)
	}
	if got := a.Allocated(); got != 1 {
		t.Fatalf("Allocated = %d, want 1", got)
	}
}

func TestAllocatorWidensAfter999(t *testing.T) {
	a := NewAllocator()
	for i := 0; i < 999; i++ {
		a.Next()
	}
	if got := a.Next(); got != "p-1000" {
		t.Fatalf("1000th ID = %q, want p-1000", got)
	}
}
