// Package engine executes scenarios: it dispatches events, performs route
// lookups, moves packets across links, and records every decision. It is the
// only package that owns network truth.
package engine

import (
	"fmt"

	"github.com/HalxDocs/packetlab/internal/eventlog"
	"github.com/HalxDocs/packetlab/internal/network"
	"github.com/HalxDocs/packetlab/internal/packet"
	"github.com/HalxDocs/packetlab/internal/sim"
	"github.com/HalxDocs/packetlab/internal/topology"
)

// ActionKind identifies a scenario action.
type ActionKind string

// ActionSendPing creates an ICMP echo request from a node.
const ActionSendPing ActionKind = "send_ping"

// Action is a scheduled user intent. The scenario loader translates file
// actions into this form so the engine has no file-format dependency.
type Action struct {
	At       sim.SimTime
	Kind     ActionKind
	From     topology.NodeID
	To       network.IPv4
	Count    int
	Interval sim.SimTime
	TTL      uint8
}

// DefaultTTL is used when an action does not set a TTL.
const DefaultTTL uint8 = 64

// DefaultEventLimit bounds a run that never drains its queue.
const DefaultEventLimit = 100_000

// Options configures an engine instance.
type Options struct {
	Actions    []Action
	EventLimit int
	Seed       uint64
}

// scheduledPing is a validated, fully expanded ping action.
type scheduledPing struct {
	At     sim.SimTime
	From   topology.NodeID
	To     network.IPv4
	TTL    uint8
	EchoID uint16
	Seq    uint16
}

// Result summarizes a run.
type Result struct {
	Processed    uint64
	Delivered    int
	Dropped      int
	Duration     sim.SimTime
	Completed    bool
	LimitReached bool
	Records      int
}

// Engine runs one scenario. It is single-threaded by design.
type Engine struct {
	topo   *topology.Topology
	opts   Options
	tables map[topology.NodeID]*network.Table
	pings  []scheduledPing

	queue *sim.Queue
	clock *sim.Clock
	log   *eventlog.Recorder
	ids   *packet.Allocator

	pingState    map[uint32]*pingState
	processed    uint64
	delivered    int
	dropped      int
	completed    bool
	limitReached bool
}

// New validates the topology and actions and prepares a run.
func New(topo *topology.Topology, opts Options) (*Engine, error) {
	if topo == nil {
		return nil, fmt.Errorf("engine: topology is nil")
	}
	if err := topo.Validate(); err != nil {
		return nil, err
	}
	if opts.EventLimit <= 0 {
		opts.EventLimit = DefaultEventLimit
	}

	tables := make(map[topology.NodeID]*network.Table, topo.NodeCount())
	for _, node := range topo.Nodes() {
		tables[node.ID] = network.NewTable(node.Routes)
	}

	pings, err := expandActions(topo, opts.Actions)
	if err != nil {
		return nil, err
	}

	e := &Engine{
		topo:      topo,
		opts:      opts,
		tables:    tables,
		pings:     pings,
		queue:     sim.NewQueue(),
		clock:     sim.NewClock(),
		log:       eventlog.NewRecorder(),
		ids:       packet.NewAllocator(),
		pingState: make(map[uint32]*pingState),
	}
	e.scheduleInitial()
	return e, nil
}

// expandActions validates actions and turns them into concrete ping events.
func expandActions(topo *topology.Topology, actions []Action) ([]scheduledPing, error) {
	var pings []scheduledPing
	for i, action := range actions {
		if action.Kind != ActionSendPing {
			return nil, fmt.Errorf("engine: action %d: unknown kind %q", i, action.Kind)
		}
		if action.At < 0 {
			return nil, fmt.Errorf("engine: action %d: time must not be negative", i)
		}
		if _, ok := topo.Node(action.From); !ok {
			return nil, fmt.Errorf("engine: action %d: unknown source node %q", i, action.From)
		}
		if !action.To.IsValid() {
			return nil, fmt.Errorf("engine: action %d: invalid destination", i)
		}
		count := action.Count
		if count == 0 {
			count = 1
		}
		if count < 0 || count > 1000 {
			return nil, fmt.Errorf("engine: action %d: count must be between 1 and 1000", i)
		}
		if action.Interval < 0 {
			return nil, fmt.Errorf("engine: action %d: interval must not be negative", i)
		}
		ttl := action.TTL
		if ttl == 0 {
			ttl = DefaultTTL
		}
		echoID := uint16(i + 1)
		for n := 0; n < count; n++ {
			pings = append(pings, scheduledPing{
				At:     action.At + action.Interval*sim.SimTime(n),
				From:   action.From,
				To:     action.To,
				TTL:    ttl,
				EchoID: echoID,
				Seq:    uint16(n + 1),
			})
		}
	}
	return pings, nil
}

// scheduleInitial resets runtime state and schedules the first events.
func (e *Engine) scheduleInitial() {
	e.clock.Reset()
	e.queue = sim.NewQueue()
	e.log.Reset()
	e.ids.Reset()
	e.pingState = make(map[uint32]*pingState)
	e.processed = 0
	e.delivered = 0
	e.dropped = 0
	e.completed = false
	e.limitReached = false

	e.record(sim.NewEvent(0, sim.RunStarted))
	for _, ping := range e.pings {
		e.schedule(ping.At, eventSendPing, ping)
	}
}

// Reset restores the initial scenario state and starts a new log epoch.
func (e *Engine) Reset() {
	e.scheduleInitial()
}

// Now returns the current virtual time.
func (e *Engine) Now() sim.SimTime { return e.clock.Now() }

// Records returns the event log for this run.
func (e *Engine) Records() []eventlog.Record { return e.log.Records() }

// Result summarizes progress so far.
func (e *Engine) Result() Result {
	return Result{
		Processed:    e.processed,
		Delivered:    e.delivered,
		Dropped:      e.dropped,
		Duration:     e.clock.Now(),
		Completed:    e.completed,
		LimitReached: e.limitReached,
		Records:      e.log.Len(),
	}
}

// Completed reports whether the queue drained and the run finished.
func (e *Engine) Completed() bool { return e.completed }

// LimitReached reports whether the engine stopped at the event limit.
func (e *Engine) LimitReached() bool { return e.limitReached }

// Step processes exactly one event.
func (e *Engine) Step() (bool, error) {
	if e.completed || e.limitReached {
		return false, nil
	}
	ev, ok := e.queue.Pop()
	if !ok {
		e.finish()
		return false, nil
	}
	if err := e.clock.AdvanceTo(ev.At); err != nil {
		return false, err
	}
	e.processed++
	if err := e.dispatch(ev); err != nil {
		return false, err
	}
	if e.processed >= uint64(e.opts.EventLimit) {
		e.limitReached = true
	}
	return true, nil
}

// Run processes events until the queue drains, the limit is reached, or
// maxEvents events have been processed. A non-positive maxEvents uses the
// configured event limit.
func (e *Engine) Run(maxEvents int) error {
	if maxEvents <= 0 {
		maxEvents = e.opts.EventLimit
	}
	for i := 0; i < maxEvents; i++ {
		if e.completed || e.limitReached {
			return nil
		}
		ok, err := e.Step()
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}
	return nil
}

// RunToCompletion runs until the queue is empty or the event limit is hit.
func (e *Engine) RunToCompletion() error { return e.Run(0) }

// dispatch routes an internal event to its handler.
func (e *Engine) dispatch(ev sim.Event) error {
	switch ev.Kind {
	case eventSendPing:
		ping, ok := ev.Payload.(scheduledPing)
		if !ok {
			return fmt.Errorf("engine: send-ping event has payload %T", ev.Payload)
		}
		return e.startPing(ping)
	case eventLinkArrive:
		arrival, ok := ev.Payload.(linkArrival)
		if !ok {
			return fmt.Errorf("engine: link-arrival event has payload %T", ev.Payload)
		}
		return e.arrive(arrival)
	default:
		return fmt.Errorf("engine: unhandled event kind %q", ev.Kind)
	}
}

// finish appends the completion record exactly once.
func (e *Engine) finish() {
	if e.completed {
		return
	}
	e.completed = true
	e.record(sim.NewEvent(e.clock.Now(), sim.RunCompleted).
		WithDetail("processed", fmt.Sprintf("%d", e.processed)).
		WithDetail("delivered", fmt.Sprintf("%d", e.delivered)).
		WithDetail("dropped", fmt.Sprintf("%d", e.dropped)))
}

// schedule pushes an internal event and returns it.
func (e *Engine) schedule(at sim.SimTime, kind sim.EventKind, payload any) sim.Event {
	return e.queue.Push(sim.Event{At: at, Kind: kind, Payload: payload})
}

// record appends an event to the log with its generated summary.
func (e *Engine) record(ev sim.Event) {
	e.log.Append(ev, eventlog.Describe(ev))
}
