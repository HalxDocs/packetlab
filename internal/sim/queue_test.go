package sim

import (
	"math/rand"
	"testing"
)

func TestQueueOrdersByTime(t *testing.T) {
	q := NewQueue()
	q.Push(NewEvent(Ms(3), PacketCreated))
	q.Push(NewEvent(Ms(1), PacketCreated))
	q.Push(NewEvent(Ms(2), PacketCreated))

	var got []SimTime
	for {
		e, ok := q.Pop()
		if !ok {
			break
		}
		got = append(got, e.At)
	}
	want := []SimTime{Ms(1), Ms(2), Ms(3)}
	if len(got) != len(want) {
		t.Fatalf("popped %d events, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("event %d at %s, want %s", i, got[i], want[i])
		}
	}
}

func TestQueueStableAtEqualTimes(t *testing.T) {
	q := NewQueue()
	for i := 0; i < 5; i++ {
		q.Push(NewEvent(Ms(1), EventKind("event."+string(rune('a'+i)))))
	}
	var got []string
	for {
		e, ok := q.Pop()
		if !ok {
			break
		}
		got = append(got, string(e.Kind))
	}
	want := []string{"event.a", "event.b", "event.c", "event.d", "event.e"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

func TestQueueAssignsMonotonicSequences(t *testing.T) {
	q := NewQueue()
	var last uint64
	for i := 0; i < 100; i++ {
		e := q.Push(NewEvent(Ms(1), PacketCreated))
		if i > 0 && e.Sequence <= last {
			t.Fatalf("sequence %d did not increase after %d", e.Sequence, last)
		}
		last = e.Sequence
	}
	if got := q.NextSequence(); got != 100 {
		t.Fatalf("NextSequence = %d, want 100", got)
	}
}

func TestQueueOverwritesCallerSequence(t *testing.T) {
	q := NewQueue()
	e := q.Push(Event{At: 0, Sequence: 999, Kind: PacketCreated})
	if e.Sequence != 0 {
		t.Fatalf("Sequence = %d, want queue-assigned 0", e.Sequence)
	}
}

func TestQueueEmpty(t *testing.T) {
	q := NewQueue()
	if _, ok := q.Pop(); ok {
		t.Fatal("Pop on empty queue returned an event")
	}
	if _, ok := q.Peek(); ok {
		t.Fatal("Peek on empty queue returned an event")
	}
	if q.Len() != 0 {
		t.Fatalf("Len = %d, want 0", q.Len())
	}
}

func TestQueuePeekDoesNotRemove(t *testing.T) {
	q := NewQueue()
	q.Push(NewEvent(Ms(2), PacketCreated))
	first, ok := q.Peek()
	if !ok || first.At != Ms(2) {
		t.Fatalf("Peek = %+v, %v", first, ok)
	}
	if q.Len() != 1 {
		t.Fatalf("Peek removed an event")
	}
}

// TestQueueRandomizedOrdering pushes many events in a shuffled order and
// asserts that the queue never returns time backward and never reuses a
// sequence number.
func TestQueueRandomizedOrdering(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	q := NewQueue()
	for i := 0; i < 2000; i++ {
		q.Push(NewEvent(Ms(int64(rng.Intn(100))), PacketCreated))
	}

	var lastTime SimTime
	var lastSeq uint64
	seen := make(map[uint64]bool, 2000)
	for {
		e, ok := q.Pop()
		if !ok {
			break
		}
		if e.At < lastTime {
			t.Fatalf("time moved backward: %s after %s", e.At, lastTime)
		}
		if seen[e.Sequence] {
			t.Fatalf("duplicate sequence %d", e.Sequence)
		}
		seen[e.Sequence] = true
		// Sequence is scheduling order, not pop order: within one timestamp
		// it must increase, but a later-scheduled earlier event may pop first.
		if e.At == lastTime && e.Sequence < lastSeq {
			t.Fatalf("equal-time sequence moved backward: %d after %d", e.Sequence, lastSeq)
		}
		if e.At != lastTime {
			lastSeq = 0
		}
		lastTime = e.At
		lastSeq = e.Sequence
	}
	if len(seen) != 2000 {
		t.Fatalf("popped %d unique events, want 2000", len(seen))
	}
}
