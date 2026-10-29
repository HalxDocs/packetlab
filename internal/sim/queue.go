package sim

import "container/heap"

// Queue is a stable min-priority queue of events ordered by (At, Sequence).
// The queue assigns sequence numbers when events are scheduled, which makes
// equal-time ordering reproducible across runs.
type Queue struct {
	items eventHeap
	next  uint64
}

// NewQueue returns an empty event queue.
func NewQueue() *Queue {
	return &Queue{items: make(eventHeap, 0, 16)}
}

// Len returns the number of scheduled events.
func (q *Queue) Len() int { return len(q.items) }

// Push schedules an event and assigns it the next sequence number. Any
// sequence value set by the caller is overwritten.
func (q *Queue) Push(e Event) Event {
	e.Sequence = q.next
	q.next++
	heap.Push(&q.items, e)
	return e
}

// Pop removes and returns the earliest event.
func (q *Queue) Pop() (Event, bool) {
	if len(q.items) == 0 {
		return Event{}, false
	}
	return heap.Pop(&q.items).(Event), true
}

// Peek returns the earliest event without removing it.
func (q *Queue) Peek() (Event, bool) {
	if len(q.items) == 0 {
		return Event{}, false
	}
	return q.items[0], true
}

// NextSequence returns the sequence number that the next scheduled event will
// receive.
func (q *Queue) NextSequence() uint64 { return q.next }

// eventHeap implements heap.Interface.
type eventHeap []Event

func (h eventHeap) Len() int { return len(h) }

func (h eventHeap) Less(i, j int) bool {
	if h[i].At != h[j].At {
		return h[i].At < h[j].At
	}
	return h[i].Sequence < h[j].Sequence
}

func (h eventHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *eventHeap) Push(x any) {
	*h = append(*h, x.(Event))
}

func (h *eventHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}
