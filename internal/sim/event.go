package sim

// EventKind is a stable identifier for a kind of simulation event. Kinds are
// part of the observable contract: tests and the UI match on them, so they
// must not be renamed without a schema version bump.
type EventKind string

// Public event kinds appear in the event log.
const (
	RunStarted       EventKind = "run.started"
	RunCompleted     EventKind = "run.completed"
	PacketCreated    EventKind = "packet.created"
	PacketReceived   EventKind = "packet.received"
	PacketForwarded  EventKind = "packet.forwarded"
	PacketDelivered  EventKind = "packet.delivered"
	PacketDropped    EventKind = "packet.dropped"
	RouteSelected    EventKind = "route.selected"
	LinkTransmit     EventKind = "link.transmit"
	LinkDelivered    EventKind = "link.delivered"
	LinkStateChanged EventKind = "link.state_changed"
	PingRequested    EventKind = "ping.requested"
	PingReply        EventKind = "ping.reply"
)

// ReasonCode is a stable, machine-readable explanation attached to a record.
// Human-readable summaries may change; reason codes may not.
type ReasonCode string

// Failure reasons.
const (
	ReasonNoRoute            ReasonCode = "no_route"
	ReasonLinkDown           ReasonCode = "link_down"
	ReasonTTLExpired         ReasonCode = "ttl_expired"
	ReasonNoInterface        ReasonCode = "no_interface"
	ReasonUnknownDestination ReasonCode = "unknown_destination"
	ReasonNodeDown           ReasonCode = "node_down"
	ReasonQueueFull          ReasonCode = "queue_full"
	ReasonLinkLoss           ReasonCode = "link_loss"
)

// Event is a scheduled cause inside the engine. Events in the queue are not
// themselves written to the log; handlers append records describing outcomes.
type Event struct {
	At       SimTime
	Sequence uint64
	Kind     EventKind
	NodeID   string
	LinkID   string
	PacketID string
	Reason   ReasonCode
	Details  map[string]string
	// Payload carries engine-internal data, such as the packet being
	// delivered. It is never serialized.
	Payload any
}

// NewEvent returns an event scheduled at the given virtual time.
func NewEvent(at SimTime, kind EventKind) Event {
	return Event{At: at, Kind: kind}
}

// WithNode returns a copy of the event tagged with a node.
func (e Event) WithNode(id string) Event {
	e.NodeID = id
	return e
}

// WithLink returns a copy of the event tagged with a link.
func (e Event) WithLink(id string) Event {
	e.LinkID = id
	return e
}

// WithPacket returns a copy of the event tagged with a packet.
func (e Event) WithPacket(id string) Event {
	e.PacketID = id
	return e
}

// WithReason returns a copy of the event carrying a reason code.
func (e Event) WithReason(r ReasonCode) Event {
	e.Reason = r
	return e
}

// WithPayload returns a copy of the event carrying engine-internal data.
func (e Event) WithPayload(p any) Event {
	e.Payload = p
	return e
}

// WithDetail returns a copy of the event with one detail entry set. The detail
// map is copied so events remain safe to share.
func (e Event) WithDetail(key, value string) Event {
	details := make(map[string]string, len(e.Details)+1)
	for k, v := range e.Details {
		details[k] = v
	}
	details[key] = value
	e.Details = details
	return e
}
