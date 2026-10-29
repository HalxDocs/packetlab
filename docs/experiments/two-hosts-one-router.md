# Experiment: two hosts, one router

## Question

Which path does a ping take, and which decisions explain it?

## Topology

```
host-a                 router-r1                 host-b
10.0.1.10/24           10.0.1.1/24               10.0.2.20/24
                       10.0.2.1/24
      link-a                     link-b
```

## Routes

| Node | Network | Next hop | Interface |
| --- | --- | --- | --- |
| host-a | 0.0.0.0/0 | 10.0.1.1 | host-a:eth0 |
| router-r1 | 10.0.1.0/24 | connected | router-r1:eth0 |
| router-r1 | 10.0.2.0/24 | connected | router-r1:eth1 |
| host-b | 0.0.0.0/0 | 10.0.2.1 | host-b:eth0 |

## Run

```bash
go run ./cmd/packetlab run scenarios/two-hosts-one-router.json
```

## What to look for

1. `packet.created` at host-a with packet `p-001`.
2. `route.selected` at host-a: default route via the router.
3. `link.transmit` and `link.delivered` for link-a.
4. `packet.received` at the router, followed by a second `route.selected`.
5. `packet.delivered` at host-b, then a new packet for the echo reply.
6. `ping.reply` at host-a with the round-trip time in virtual milliseconds.

## Variations

- Remove host-a's default route. The first decision becomes
  `packet.dropped reason=no_route`.
- Bring link-b down. The router selects a route, then drops the packet with
  `reason=link_down`.
- Set the ping TTL to 1. The router drops the packet with
  `reason=ttl_expired`.
- Increase both link delays to 20 ms. The reply arrives at 80 ms.

## Why this is a regression test

The scenario and its expected trace are checked in. If forwarding, ordering, or
summaries change, the golden test fails and the diff shows exactly which event
changed.
