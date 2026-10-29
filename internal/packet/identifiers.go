// Package packet defines packet identity and the packet model used by the
// simulation. Identity is stable across every hop so one packet can be traced
// end to end.
package packet

import "fmt"

// ID uniquely identifies a packet within a run. IDs are allocated in order and
// never reused, even after a packet is dropped.
type ID string

// Allocator produces deterministic packet IDs.
type Allocator struct {
	next uint64
}

// NewAllocator returns an allocator positioned before the first ID.
func NewAllocator() *Allocator { return &Allocator{} }

// Next returns the next packet ID, formatted as p-001.
func (a *Allocator) Next() ID {
	a.next++
	return ID(fmt.Sprintf("p-%03d", a.next))
}

// Reset returns the allocator to its initial state. IDs from a previous run
// may be reused only after a reset, which starts a new log epoch.
func (a *Allocator) Reset() { a.next = 0 }

// Allocated returns how many IDs have been issued.
func (a *Allocator) Allocated() uint64 { return a.next }
