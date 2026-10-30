# PacketLab web workbench

A React + TypeScript trace viewer for PacketLab. It **replays** precomputed
engine output; it never simulates a packet or invents an event.

## Run it

From the repository root:

```bash
make traces          # build the CLI and write web/public/traces/*.json
cd web
npm install
npm run dev          # http://localhost:5173
```

Production build:

```bash
npm run typecheck
npm test -- --run
npm run build
```

## Data flow

```
scenarios/*.json
      │
      ▼
packetlab trace
      │
      ▼
web/public/traces/<scenario-id>.json
      │
      ▼
fetch → replay cursor → topology + timeline + inspector
```

`make traces` regenerates every bundle. Re-run it whenever a scenario or the
engine's event log changes, then commit the updated bundles. The viewer is a
static site: the JSON files are the only runtime data source.

## Interface

- **Topology** — a deterministic left-to-right BFS layout. Blue marks the
  current node, link, or selected event; a down link is dashed; a drop is red.
  The packet token sits where the current record says it is. Nothing animates
  independently of the event log.
- **Event log** — every record in order, with filters for packets, routing,
  links, and failures. The replay cursor and the selected row are the same
  state.
- **Inspector** — the current record's inputs, the full journey of its packet,
  and the interfaces and static routes of the node involved.
- **Replay controls** — first, previous, play/pause, next, last, speed, and a
  scrubber. Playback only moves a cursor; virtual time comes from the records.

## Design rules

The palette is paper, ink, and hairlines. Blue (`#2563eb` / `#1d4ed8`) is the
only interaction accent. Green is reserved for successful outcomes (delivery,
echo reply, link up). Red appears only for a failure reason, so a dropped packet
reads as a drop rather than decoration. The workbench is desktop-first and
assumes a 1120 px minimum width.
