package eventlog

import (
	"testing"

	"github.com/HalxDocs/packetlab/internal/sim"
)

func TestRecorderAssignsIndexes(t *testing.T) {
	r := NewRecorder()
	for i := 0; i < 3; i++ {
		rec := r.Append(sim.NewEvent(sim.Ms(int64(i)), sim.PacketCreated), "created packet")
		if rec.Index != uint64(i) {
			t.Fatalf("record %d index = %d", i, rec.Index)
		}
	}
	if r.Len() != 3 {
		t.Fatalf("Len = %d, want 3", r.Len())
	}
}

func TestRecorderCopiesDetails(t *testing.T) {
	r := NewRecorder()
	e := sim.NewEvent(0, sim.PacketCreated).WithDetail("a", "1")
	rec := r.Append(e, "created packet")
	rec.Details["a"] = "mutated"

	if got := r.Records()[0].Details["a"]; got != "1" {
		t.Fatalf("stored detail = %q, want 1", got)
	}
}

func TestRecorderRecordsIsCopy(t *testing.T) {
	r := NewRecorder()
	r.Append(sim.NewEvent(0, sim.PacketCreated), "created packet")
	records := r.Records()
	records[0].Summary = "changed"

	if got := r.Records()[0].Summary; got != "created packet" {
		t.Fatalf("stored summary = %q", got)
	}
}

func TestRecorderReset(t *testing.T) {
	r := NewRecorder()
	r.Append(sim.NewEvent(0, sim.PacketCreated), "created packet")
	r.Reset()
	if r.Len() != 0 {
		t.Fatalf("Len after Reset = %d", r.Len())
	}
	rec := r.Append(sim.NewEvent(0, sim.PacketCreated), "created packet")
	if rec.Index != 0 {
		t.Fatalf("index after Reset = %d, want 0", rec.Index)
	}
}
