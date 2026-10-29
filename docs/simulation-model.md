# Simulation model

PacketLab uses discrete-event simulation. The engine does not tick every
component on a fixed interval. It keeps a priority queue of future events and
jumps virtual time directly to the next one.

## Virtual time

`sim.SimTime` is an `int64` count of nanoseconds since the start of a run. It is
independent of wall-clock time. A packet scheduled over a 2 ms link produces a
delivery event at `t + 2ms`; the engine processes it immediately, without
sleeping for two real milliseconds.

Virtual time never moves backward. `Clock.AdvanceTo` rejects earlier
timestamps, and the queue guarantees non-decreasing event times.

## Events

```go
type Event struct {
    At       SimTime
    Sequence uint64
    Kind     EventKind
    NodeID   string
    LinkID   string
    PacketID string
    Reason   ReasonCode
    Details  map[string]string
    Payload  any // engine-internal; never serialized
}
```

`Sequence` is assigned by the queue at scheduling time. It is the tie-breaker
for events that share a timestamp, which makes equal-time ordering stable
across runs.

Events in the queue are internal causes. The log records outcomes. For example,
the queue holds `engine.link_arrive`, and the handler appends `link.delivered`
and then `packet.received`.

## Event loop

```
while queue is not empty and running:
    event = queue.pop_earliest()
    clock.advance_to(event.at)
    dispatch(event)
    if event limit reached: stop
```

There is one goroutine. No locks are required, and no ordering depends on the
Go scheduler.

## Packet lifecycle

```
create ──► route ──► transmit ──► arrive ──► forward ──► deliver
              │                       │
              └── no route ──► drop    └── TTL expired ──► drop
```

1. A host creates a packet with a stable ID and a TTL.
2. The host looks up a route for the destination.
3. If the route exists, the packet is serialized onto the egress link and
   scheduled to arrive after serialization plus propagation delay.
4. A router receives the packet, decrements TTL, and looks up a route.
5. The packet crosses links until it reaches a host that owns the destination
   address, or is dropped with a reason.

Packet identity does not change across hops. Every record that concerns the
packet carries the same `packetId`, which makes end-to-end tracing possible.

## Link model

| Property | Behavior |
| --- | --- |
| Propagation delay | Added to the arrival timestamp after serialization |
| Bandwidth | `serialization = ceil(size_bits × 1e9 / bps)` ns; 0 means unlimited |
| State | A down link refuses transmission with `reason=link_down` |
| Queue limit | Reserved for phase 2; the field is validated but not enforced |

Packet size is 28 bytes of IPv4 + ICMP header plus payload. This is documented
as a simplification: PacketLab does not model Ethernet framing, preamble, or
inter-frame gaps.

## Failure outcomes

Every failure is a record with a stable reason code:

| Reason | Meaning |
| --- | --- |
| `no_route` | No route matched the destination |
| `link_down` | The selected egress link is down |
| `ttl_expired` | TTL reached zero at a router |
| `no_interface` | The route named an interface that has no link |
| `unknown_destination` | A host received a packet for an address it does not own |
| `node_down` | The node is administratively down |

## Determinism test

The engine is run twice from the same scenario and the normalized records are
compared byte for byte. The golden test does the same against a checked-in
trace, so an unexpected change in ordering, summaries, or reasons fails CI.

## Deliberate omissions

- No concurrency in the core.
- No probabilistic loss in milestone 1.
- No fragmentation, ARP, or Ethernet.
- No retransmission; ICMP echo is a single request/reply exchange.
