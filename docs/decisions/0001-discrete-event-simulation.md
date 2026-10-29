# ADR 0001: Discrete-event simulation core

- Status: accepted
- Date: 2026-10-29

## Context

PacketLab must demonstrate delay, queueing, timeouts, and failure recovery. A
fixed-tick simulation advances every component on every tick, wastes work, and
makes event ordering depend on tick boundaries. It also encourages animations
that are not tied to exact event times.

## Decision

Use a discrete-event simulation:

- A virtual clock stores the current `SimTime`.
- A min-priority queue holds future events.
- The engine pops the earliest event, advances the clock directly to its
  timestamp, and dispatches it.
- Handlers schedule future events rather than sleeping.

## Consequences

Positive:

- Virtual time is exact and inspectable.
- Simulations run as fast as the CPU allows.
- Event ordering is explicit and testable.
- The UI can render a deterministic timeline from timestamps.

Negative:

- Handlers must be written as state machines; there is no "update everything"
  loop.
- Every future action must be representable as an event.
- Debugging requires reading the event log rather than stepping a tick loop.

## Alternatives considered

- **Fixed 1 ms tick.** Simple, but slow, imprecise, and nondeterministic at tick
  boundaries.
- **Real-time execution with sleeps.** Ties behavior to wall-clock scheduling
  and makes tests slow and flaky.
