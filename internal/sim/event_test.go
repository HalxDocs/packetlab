package sim

import "testing"

func TestEventBuilders(t *testing.T) {
	e := NewEvent(Ms(1), PacketCreated).
		WithNode("host-a").
		WithLink("link-a").
		WithPacket("p-001").
		WithReason(ReasonNoRoute).
		WithDetail("src", "10.0.1.10").
		WithDetail("dst", "10.0.2.20")

	if e.At != Ms(1) {
		t.Errorf("At = %s, want 1ms", e.At)
	}
	if e.NodeID != "host-a" || e.LinkID != "link-a" || e.PacketID != "p-001" {
		t.Errorf("identifiers not applied: %+v", e)
	}
	if e.Reason != ReasonNoRoute {
		t.Errorf("Reason = %q, want %q", e.Reason, ReasonNoRoute)
	}
	if got := e.Details["src"]; got != "10.0.1.10" {
		t.Errorf("src detail = %q", got)
	}
	if len(e.Details) != 2 {
		t.Errorf("len(Details) = %d, want 2", len(e.Details))
	}
}

func TestEventDetailIsCopied(t *testing.T) {
	base := NewEvent(0, PacketCreated).WithDetail("a", "1")
	derived := base.WithDetail("b", "2")

	if _, ok := base.Details["b"]; ok {
		t.Fatal("WithDetail mutated the original event")
	}
	if derived.Details["a"] != "1" || derived.Details["b"] != "2" {
		t.Fatalf("derived details = %v", derived.Details)
	}
}

func TestEventPayload(t *testing.T) {
	type payload struct{ N int }
	e := NewEvent(0, PacketCreated).WithPayload(payload{N: 7})
	got, ok := e.Payload.(payload)
	if !ok || got.N != 7 {
		t.Fatalf("Payload = %#v", e.Payload)
	}
}
