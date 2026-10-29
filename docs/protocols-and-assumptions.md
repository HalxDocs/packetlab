# Protocols and assumptions

PacketLab is a teaching model, not a standards-compliant stack. Every
simplification below is intentional and should be visible to users.

## IPv4-like addressing

- Addresses are 32-bit IPv4 values parsed with `net/netip`.
- Interfaces carry an address and prefix length, for example `10.0.1.10/24`.
- The network address is derived by masking; the host address is preserved.
- There is no ARP. A directly connected next hop is assumed reachable at the
  link layer.
- There is no fragmentation. Packets are small enough for every link.

## Routing

- Each node owns a static route table.
- Lookup uses longest-prefix match.
- Ties are broken deterministically: lower metric, then network string, then
  next hop, then interface.
- A route names its egress interface explicitly. There is no recursive
  next-hop resolution.
- A next hop of `0.0.0.0` (or an omitted next hop) means "directly connected".
- There are no dynamic routing protocols.

## Forwarding

- Only nodes of kind `router` forward.
- Routers decrement TTL before forwarding. A packet whose TTL is 1 is dropped
  with `ttl_expired`; the model does not yet generate an ICMP time-exceeded
  message.
- Hosts accept a packet only when the destination address matches one of their
  interfaces.
- A host that receives an ICMP echo request generates a reply to the source
  address using its own route table.

## ICMP

The echo format is ICMP-like, not wire-compatible:

```
byte 0      type (8 = request, 0 = reply)
bytes 1-2   identifier
bytes 3-4   sequence
bytes 5..   opaque payload
```

There is no checksum, no code byte, and no other ICMP message type. The engine
labels these packets as `icmp` and records `ping.reply` with the measured
round-trip time in virtual milliseconds.

## Links

- Propagation delay is fixed per link.
- Bandwidth produces a serialization delay. Zero bandwidth means unlimited.
- Links are point-to-point. There are no shared media or switches yet.
- A down link rejects transmission; the packet is dropped with `link_down`.
- Queue capacity and probabilistic loss are modeled in phase 2.

## Time

- Virtual time is nanoseconds.
- There is no wall-clock dependency in the core.
- Scenario files express time in integer milliseconds.

## What this model cannot tell you

- How a real kernel would behave under the same topology.
- How TCP congestion control reacts (TCP arrives in phase 5 as a teaching
  subset).
- How hardware queues, offloading, or driver behavior affect latency.
- Whether a real network would drop the same packet. PacketLab is a controlled
  model with explicit rules.

## Versioning

Scenario files carry a `schemaVersion`. The current version is `1.0`. Public
event kinds and reason codes are stable identifiers; human-readable summaries
may change without a schema bump.
