// Package sim contains the deterministic discrete-event core: virtual time,
// events, a stable priority queue, and the virtual clock. It has no
// dependency on networking, transport, or presentation code.
package sim

import (
	"strconv"
)

// SimTime is a point or duration in virtual time, stored as nanoseconds since
// the start of a run. Virtual time is independent of wall-clock time: the
// engine jumps directly to the timestamp of the next scheduled event.
type SimTime int64

const (
	Nanosecond  SimTime = 1
	Microsecond SimTime = 1000 * Nanosecond
	Millisecond SimTime = 1000 * Microsecond
	Second      SimTime = 1000 * Millisecond
	Minute      SimTime = 60 * Second
)

// Ms returns n milliseconds of virtual time.
func Ms(n int64) SimTime { return SimTime(n) * Millisecond }

// Us returns n microseconds of virtual time.
func Us(n int64) SimTime { return SimTime(n) * Microsecond }

// Add returns t+d.
func (t SimTime) Add(d SimTime) SimTime { return t + d }

// Sub returns t-d. The result is negative when d is later than t.
func (t SimTime) Sub(d SimTime) SimTime { return t - d }

// Before reports whether t is earlier than d.
func (t SimTime) Before(d SimTime) bool { return t < d }

// Milliseconds returns t as fractional milliseconds.
func (t SimTime) Milliseconds() float64 {
	return float64(t) / float64(Millisecond)
}

// String renders virtual time with the most readable unit.
func (t SimTime) String() string {
	switch {
	case t < 0:
		return "-" + (-t).String()
	case t == 0:
		return "0ms"
	case t < Millisecond:
		return format(float64(t)/float64(Microsecond)) + "µs"
	case t < Second:
		return format(float64(t)/float64(Millisecond)) + "ms"
	default:
		return format(float64(t)/float64(Second)) + "s"
	}
}

// TraceString renders virtual time in milliseconds with three decimal places.
// It is the canonical timestamp format for traces and the CLI.
func (t SimTime) TraceString() string {
	return strconv.FormatFloat(t.Milliseconds(), 'f', 3, 64) + "ms"
}

func format(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
