package sim

import "fmt"

// Clock tracks the current virtual time. It refuses to move backward, which
// protects the monotonic-time invariant at the lowest layer.
type Clock struct {
	now SimTime
}

// NewClock returns a clock positioned at virtual time zero.
func NewClock() *Clock { return &Clock{} }

// Now returns the current virtual time.
func (c *Clock) Now() SimTime { return c.now }

// AdvanceTo moves the clock forward to t. Moving to the current time is
// allowed; moving backward is an error.
func (c *Clock) AdvanceTo(t SimTime) error {
	if t < c.now {
		return fmt.Errorf("sim: virtual time cannot move backward: %s -> %s", c.now, t)
	}
	c.now = t
	return nil
}

// Reset returns the clock to virtual time zero.
func (c *Clock) Reset() { c.now = 0 }
