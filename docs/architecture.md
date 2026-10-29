# Architecture

PacketLab separates the simulation core from transport and presentation. The Go
engine is authoritative; the web interface is a renderer.

```
┌──────────────────────────────┐        commands        ┌──────────────────────────────┐
│  React + TypeScript (web)    │ ─────────────────────► │  Go simulation engine        │
│  canvas · timeline · inspector│ ◄───────────────────── │  topology · clock · routing  │
└──────────────────────────────┘        events          └──────────────────────────────┘
                                                                    │
                                                          append-only event log
```

## Layers

| Layer | Package | Responsibility |
| --- | --- | --- |
| Core time and events | `internal/sim` | `SimTime`, event kinds, stable priority queue, virtual clock |
| Addressing | `internal/network` | IPv4 values, prefixes, subnet matching, route tables |
| Topology | `internal/topology` | Nodes, interfaces, links, structural validation |
| Packets | `internal/packet` | Stable packet identity, protocol numbers, TTL |
| Protocols | `internal/protocols/icmp` | Echo request/reply encoding |
| Event log | `internal/eventlog` | Structured records, summaries, JSONL, table output |
| Engine | `internal/engine` | Event dispatch, routing decisions, link transmission, delivery |
| Scenarios | `internal/scenario` | Versioned JSON input, decoding, validation, topology building |
| Session | `internal/session` | `Run`, `Pause`, `Step`, `Reset` command surface |
| CLI | `internal/cli` | Headless execution and trace output |

Dependencies point downward. `sim` and `network` depend on nothing but the
standard library. The engine depends on every model package; no model package
depends on the engine. Nothing in the core imports HTTP, WebSocket, or React.

## Determinism boundary

The engine is single-threaded. Determinism depends on four rules:

1. Events are ordered by `(SimTime, Sequence)`; the sequence number is assigned
   when the event is scheduled.
2. No map iteration influences behavior or output. Anything derived from a map
   is sorted first.
3. No wall-clock calls (`time.Now`, timers, sleeps) inside the core.
4. Randomness, when introduced, comes from a seeded source owned by the
   session, never from the global `math/rand` generator.

## Ownership of state

The frontend never mutates network state directly. Every change is a command
that the session validates and applies:

```
CreateSession · ValidateTopology · Run · Pause · Step · Reset
SetLinkState · UpdateLinkModel · SendPing
```

Milestone 1 implements `Run`, `Step`, `Reset`, and `SendPing` through scenario
actions. `Pause` and link mutation arrive with the interactive session layer.

## Event flow

```
app action ──► engine schedules event ──► queue (time, seq)
                                              │
                          ┌───────────────────┘
                          ▼
                 pop earliest ──► advance clock ──► handler
                                                     │
                                    ┌────────────────┴───────────────┐
                                    ▼                                ▼
                            append record(s)                 schedule next event
```

A scheduled transmission is not a delivery. The log distinguishes
`link.transmit` (attempted, packet enters the link) from `link.delivered`
(completed, packet reaches the far interface).

## Transport

The CLI is the first transport. A future HTTP/WebSocket adapter will reuse the
same `session.Session` API and stream `eventlog.Record` values to the browser.
Schemas are versioned so saved scenarios and recorded traces stay readable.

## Related documents

- [Simulation model](simulation-model.md)
- [Protocols and assumptions](protocols-and-assumptions.md)
- [ADR 0001: discrete-event simulation](decisions/0001-discrete-event-simulation.md)
- [ADR 0002: engine owns network truth](decisions/0002-engine-owns-network-truth.md)
- [ADR 0003: stable event ordering](decisions/0003-stable-event-ordering.md)
