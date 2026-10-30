# PacketLab

An interactive laboratory for understanding how computer networks actually work.

PacketLab is a deterministic, discrete-event network simulator. A Go engine owns
network truth — topology, addressing, routes, links, virtual time, and packet
state — and emits a structured event for every decision it makes. A React +
TypeScript interface replays that event stream on a topology canvas, an event
timeline, and a packet inspector.

The goal is a flight simulator for networks: build a small topology, send
packets, introduce failures, and replay exactly what happened.

```
HOST A                 ROUTER R1                 HOST B
10.0.1.10    ────────  10.0.1.1 / 10.0.2.1  ────  10.0.2.20
             link-a                       link-b
```

## Status

Milestone 1 — a headless Go simulator that moves one packet from Host A through
Router R1 to Host B and records every decision. The web interface is a trace
viewer over real engine output; it does not simulate anything itself.

## Quick start

Requires Go 1.26 or newer. The web workbench also requires Node 22 or newer.

```bash
git clone https://github.com/HalxDocs/packetlab.git
cd packetlab

# Run the canonical scenario and print the event trace.
go run ./cmd/packetlab run scenarios/two-hosts-one-router.json

# Machine-readable event log.
go run ./cmd/packetlab run --json scenarios/two-hosts-one-router.json

# Validate a scenario without running it.
go run ./cmd/packetlab validate scenarios/two-hosts-one-router.json

# Run the test suite.
go test ./...
```

Expected trace:

```
0.000ms  packet.created    host-a       p-001  created ICMP echo request
0.000ms  route.selected    host-a       p-001  selected 0.0.0.0/0 via 10.0.1.1 on host-a:eth0
1.000ms  link.delivered    router-r1    p-001  arrived on router-r1:eth0
...
2.000ms  packet.delivered  host-b       p-001  delivered to host-b
```

## Web workbench

The viewer replays the engine's event log on a topology canvas, a filterable
timeline, and a packet inspector. It does not simulate anything itself, so the
trace bundles must be generated first:

```bash
make traces
cd web
npm install
npm run dev
```

Open `http://localhost:5173`. Re-run `make traces` after changing a scenario or
the engine. See [web/README.md](web/README.md) for the build, test, and design
details.

## What PacketLab models

- A virtual clock and a stable priority queue of events ordered by
  `(timestamp, sequence)`.
- IPv4-like addressing, prefix matching, route tables, longest-prefix match,
  next-hop selection, and TTL handling.
- Links with propagation delay, bandwidth-derived serialization delay, and
  up/down state.
- ICMP-style echo request/reply.
- Structured, append-only event records with stable kinds and reason codes.

## What PacketLab does not model (yet)

- A production router or a complete TCP/IP stack.
- Live packet capture, raw sockets, or real traffic.
- Ethernet framing, ARP, fragmentation, or congestion control.
- Probabilistic loss (the seeded loss model is phase 2).

See [docs/protocols-and-assumptions.md](docs/protocols-and-assumptions.md) for
the full list of simplifications.

## Repository layout

```
cmd/packetlab          CLI entry point
internal/sim           virtual time, events, queue, clock
internal/topology      nodes, interfaces, links, validation
internal/network       IPv4, prefixes, route tables
internal/packet        packet model, identifiers, TTL
internal/protocols     protocol handlers (ICMP)
internal/eventlog      structured records, JSONL, table output
internal/engine        event dispatch, forwarding, delivery
internal/scenario      versioned scenario files
internal/session       run/pause/step/reset command surface
internal/cli           command implementations
scenarios              canonical, versioned experiments
web                    React + TypeScript trace viewer
```

## Design principles

- **Correctness before spectacle.** A beautiful canvas cannot compensate for
  incorrect packet behavior.
- **Every event explains itself.** Record the decision, its inputs, and a
  stable reason code.
- **Reproducibility is a feature.** The same scenario and seed produce the same
  ordered event log.
- **Separate model from view.** The engine is authoritative; the frontend
  renders.
- **Small protocol slices.** Ship a complete, testable path before adding
  breadth.

## Documentation

- [Architecture](docs/architecture.md)
- [Simulation model](docs/simulation-model.md)
- [Protocols and assumptions](docs/protocols-and-assumptions.md)
- [Decisions](docs/decisions/)

## License

MIT
