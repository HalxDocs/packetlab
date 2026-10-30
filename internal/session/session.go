// Package session exposes the command surface used by transports: run, pause,
// step, and reset. It wraps the engine and tracks lifecycle state.
package session

import (
	"github.com/HalxDocs/packetlab/internal/engine"
	"github.com/HalxDocs/packetlab/internal/eventlog"
	"github.com/HalxDocs/packetlab/internal/sim"
	"github.com/HalxDocs/packetlab/internal/topology"
)

// State is the lifecycle state of a session.
type State string

const (
	// StateReady means the session has not processed any events.
	StateReady State = "ready"
	// StateRunning means events are being processed.
	StateRunning State = "running"
	// StatePaused means processing stopped before completion.
	StatePaused State = "paused"
	// StateCompleted means the event queue drained.
	StateCompleted State = "completed"
	// StateLimit means the engine stopped at the event limit.
	StateLimit State = "limit"
)

// Session owns one engine and its command lifecycle.
type Session struct {
	scenarioID string
	engine     *engine.Engine
	state      State
	pause      bool
}

// New creates a session for a built topology.
func New(scenarioID string, topo *topology.Topology, opts engine.Options) (*Session, error) {
	e, err := engine.New(topo, opts)
	if err != nil {
		return nil, err
	}
	return &Session{scenarioID: scenarioID, engine: e, state: StateReady}, nil
}

// ScenarioID returns the scenario identifier.
func (s *Session) ScenarioID() string { return s.scenarioID }

// State returns the current lifecycle state.
func (s *Session) State() State { return s.state }

// Run processes events until completion, the limit, or a pause request.
// A non-positive maxEvents uses the engine's configured limit.
func (s *Session) Run(maxEvents int) error {
	if s.state == StateCompleted || s.state == StateLimit {
		return nil
	}
	s.state = StateRunning
	s.pause = false
	for i := 0; maxEvents <= 0 || i < maxEvents; i++ {
		if s.pause {
			s.pause = false
			s.state = StatePaused
			return nil
		}
		ok, err := s.engine.Step()
		if err != nil {
			s.state = StatePaused
			return err
		}
		if !ok {
			s.state = s.terminalState()
			return nil
		}
	}
	s.state = s.terminalState()
	return nil
}

// Step processes exactly one event.
func (s *Session) Step() (bool, error) {
	if s.state == StateCompleted || s.state == StateLimit {
		return false, nil
	}
	if s.state == StateReady {
		s.state = StatePaused
	}
	ok, err := s.engine.Step()
	if err != nil {
		return false, err
	}
	if !ok {
		s.state = s.terminalState()
		return false, nil
	}
	if s.engine.Completed() || s.engine.LimitReached() {
		s.state = s.terminalState()
	}
	return true, nil
}

// Pause requests that a running loop stop at the next safe boundary. It is a
// no-op unless the session is currently running.
func (s *Session) Pause() {
	if s.state == StateRunning {
		s.pause = true
	}
}

// Reset restores the initial scenario and starts a new log epoch.
func (s *Session) Reset() {
	s.engine.Reset()
	s.state = StateReady
	s.pause = false
}

// Records returns the current event log.
func (s *Session) Records() []eventlog.Record { return s.engine.Records() }

// Result summarizes the run.
func (s *Session) Result() engine.Result { return s.engine.Result() }

// Now returns the current virtual time.
func (s *Session) Now() sim.SimTime { return s.engine.Now() }

func (s *Session) terminalState() State {
	switch {
	case s.engine.Completed():
		return StateCompleted
	case s.engine.LimitReached():
		return StateLimit
	default:
		return StatePaused
	}
}

// String implements fmt.Stringer.
func (s State) String() string { return string(s) }

// Ensure the state is one of the known values.
func (s State) Valid() bool {
	switch s {
	case StateReady, StateRunning, StatePaused, StateCompleted, StateLimit:
		return true
	default:
		return false
	}
}
