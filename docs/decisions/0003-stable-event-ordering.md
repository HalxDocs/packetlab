# ADR 0003: Stable event ordering

- Status: accepted
- Date: 2026-10-29

## Context

Two packets can be scheduled for the same virtual millisecond. If their order
depends on Go map iteration, goroutine scheduling, or insertion into an
unsorted structure, the same scenario can produce different event logs. That
breaks replay, golden tests, and bug reports.

## Decision

Order events by `(At, Sequence)`:

- `At` is the virtual timestamp.
- `Sequence` is a monotonically increasing counter assigned by the queue when
  an event is scheduled.
- The queue is a binary min-heap with a strict comparator on those two fields.

No behavior may depend on map iteration. Any output derived from a map is
sorted first. The simulation core runs on a single goroutine.

## Consequences

Positive:

- Reruns are byte-identical.
- Equal-time events have a documented, reproducible order.
- Golden traces can be checked into the repository.

Negative:

- Scheduling order becomes part of the observable contract. Reordering two
  `Schedule` calls can change a golden trace.
- Sequence numbers are per-run, not globally meaningful.

## Alternatives considered

- **Order by kind priority at equal timestamps.** Tempting, but it hides real
  scheduling order and makes handlers harder to reason about.
- **Allow nondeterministic ties and normalize in tests.** Rejected: replay
  would still differ for users.
