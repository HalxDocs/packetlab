package eventlog

import (
	"github.com/HalxDocs/packetlab/internal/sim"
)

// Recorder accumulates records for one run. It is append-only; Reset starts a
// new epoch and is the only way to clear it.
type Recorder struct {
	records []Record
}

// NewRecorder returns an empty recorder.
func NewRecorder() *Recorder { return &Recorder{} }

// Append adds a record for an event. The index is assigned by the recorder.
func (r *Recorder) Append(e sim.Event, summary string) Record {
	rec := Record{
		Index:    uint64(len(r.records)),
		At:       e.At,
		Kind:     e.Kind,
		NodeID:   e.NodeID,
		LinkID:   e.LinkID,
		PacketID: e.PacketID,
		Reason:   e.Reason,
		Summary:  summary,
	}
	if len(e.Details) > 0 {
		rec.Details = make(map[string]string, len(e.Details))
		for k, v := range e.Details {
			rec.Details[k] = v
		}
	}
	r.records = append(r.records, rec)
	return cloneRecord(rec)
}

// cloneRecord returns a record whose detail map is independent of the stored
// entry, so callers cannot mutate the log through a returned record.
func cloneRecord(rec Record) Record {
	if len(rec.Details) == 0 {
		return rec
	}
	details := make(map[string]string, len(rec.Details))
	for k, v := range rec.Details {
		details[k] = v
	}
	rec.Details = details
	return rec
}

// Records returns a copy of the recorded entries.
func (r *Recorder) Records() []Record {
	return append([]Record(nil), r.records...)
}

// Len returns the number of records.
func (r *Recorder) Len() int { return len(r.records) }

// Reset clears the recorder and starts a new epoch.
func (r *Recorder) Reset() { r.records = r.records[:0] }
