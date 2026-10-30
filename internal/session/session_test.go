package session

import (
	"testing"

	"github.com/HalxDocs/packetlab/internal/scenario"
)

func newTestSession(t *testing.T) *Session {
	t.Helper()
	s, err := scenario.Load("../../scenarios/two-hosts-one-router.json")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	topo, opts, err := s.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	sess, err := New(s.ID, topo, opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return sess
}

func TestSessionRunsToCompletion(t *testing.T) {
	s := newTestSession(t)
	if s.State() != StateReady {
		t.Fatalf("initial state = %q, want ready", s.State())
	}
	if err := s.Run(0); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if s.State() != StateCompleted {
		t.Fatalf("state = %q, want completed", s.State())
	}
	result := s.Result()
	if result.Delivered != 2 || result.Dropped != 0 {
		t.Fatalf("result = %+v", result)
	}
	if len(s.Records()) != 25 {
		t.Fatalf("records = %d, want 25", len(s.Records()))
	}
}

func TestSessionStep(t *testing.T) {
	s := newTestSession(t)
	ok, err := s.Step()
	if err != nil {
		t.Fatalf("Step: %v", err)
	}
	if !ok {
		t.Fatal("Step returned no event")
	}
	if s.State() != StatePaused {
		t.Fatalf("state = %q, want paused after a single step", s.State())
	}
	if got := s.Result().Processed; got != 1 {
		t.Fatalf("Processed = %d, want 1", got)
	}

	if err := s.Run(0); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if s.State() != StateCompleted {
		t.Fatalf("state = %q, want completed", s.State())
	}
}

func TestSessionStepAfterCompletion(t *testing.T) {
	s := newTestSession(t)
	if err := s.Run(0); err != nil {
		t.Fatalf("Run: %v", err)
	}
	ok, err := s.Step()
	if err != nil {
		t.Fatalf("Step: %v", err)
	}
	if ok {
		t.Fatal("Step after completion processed an event")
	}
}

func TestSessionReset(t *testing.T) {
	s := newTestSession(t)
	if err := s.Run(0); err != nil {
		t.Fatalf("Run: %v", err)
	}
	s.Reset()
	if s.State() != StateReady {
		t.Fatalf("state = %q, want ready", s.State())
	}
	if got := len(s.Records()); got != 1 {
		t.Fatalf("records after Reset = %d, want 1", got)
	}
	if got := s.Result().Processed; got != 0 {
		t.Fatalf("Processed after Reset = %d", got)
	}
	if err := s.Run(0); err != nil {
		t.Fatalf("Run after Reset: %v", err)
	}
	if s.State() != StateCompleted {
		t.Fatalf("state = %q", s.State())
	}
}

func TestSessionPauseIsNoOpWhenReady(t *testing.T) {
	s := newTestSession(t)
	s.Pause()
	if s.State() != StateReady {
		t.Fatalf("state = %q, want ready", s.State())
	}
}

func TestSessionScenarioID(t *testing.T) {
	s := newTestSession(t)
	if got := s.ScenarioID(); got != "two-hosts-one-router" {
		t.Fatalf("ScenarioID = %q", got)
	}
}

func TestStateValid(t *testing.T) {
	for _, state := range []State{StateReady, StateRunning, StatePaused, StateCompleted, StateLimit} {
		if !state.Valid() {
			t.Errorf("%q should be valid", state)
		}
	}
	if State("bogus").Valid() {
		t.Error("bogus state should be invalid")
	}
}
