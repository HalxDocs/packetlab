# ADR 0002: The Go engine owns network truth

- Status: accepted
- Date: 2026-10-29

## Context

A network visualizer can drift into letting the frontend animate packets along
links independently of any model. That produces a convincing demo that teaches
the wrong thing: the animation becomes the source of truth, and failures are
cosmetic.

## Decision

The Go engine is authoritative for topology, addressing, routes, link state,
packet state, and time. The frontend:

- submits commands (`Run`, `Pause`, `Step`, `Reset`, `SendPing`, ...),
- renders structured events emitted by the engine,
- never advances a packet on its own.

The web UI in this repository is a replay viewer over real engine output. Its
playback controls move through recorded events; they do not simulate.

## Consequences

Positive:

- Every visible packet movement corresponds to an engine decision.
- The same engine runs headlessly in tests and the CLI.
- The UI can be replaced without touching network behavior.

Negative:

- The frontend needs an event schema and a transport.
- UI work cannot start until the engine emits meaningful events.
- Animations must interpolate between timestamps rather than run on a timer
  loop with independent state.

## Enforcement

- No simulation package imports HTTP, WebSocket, or UI code.
- Scenario actions and session commands are the only ways to change state.
- Determinism tests compare engine output, not rendered output.
