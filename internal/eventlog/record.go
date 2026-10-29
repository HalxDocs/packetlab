// Package eventlog turns simulation events into append-only, structured
// records. Records power the CLI trace, the web timeline, debugging, replay,
// and test assertions.
package eventlog

import (
	"encoding/json"

	"github.com/HalxDocs/packetlab/internal/sim"
)

// Record is one immutable entry in a run's event log.
type Record struct {
	Index    uint64
	At       sim.SimTime
	Kind     sim.EventKind
	NodeID   string
	LinkID   string
	PacketID string
	Reason   sim.ReasonCode
	Summary  string
	Details  map[string]string
}

// wireRecord is the JSON representation. Timestamps are exposed both as
// nanoseconds (exact) and milliseconds (convenient for the UI).
type wireRecord struct {
	Index    uint64            `json:"index"`
	AtMs     float64           `json:"atMs"`
	AtNs     int64             `json:"atNs"`
	Kind     string            `json:"kind"`
	NodeID   string            `json:"nodeId,omitempty"`
	LinkID   string            `json:"linkId,omitempty"`
	PacketID string            `json:"packetId,omitempty"`
	Reason   string            `json:"reason,omitempty"`
	Summary  string            `json:"summary"`
	Details  map[string]string `json:"details,omitempty"`
}

// MarshalJSON encodes a record with a stable field order.
func (r Record) MarshalJSON() ([]byte, error) {
	return json.Marshal(wireRecord{
		Index:    r.Index,
		AtMs:     r.At.Milliseconds(),
		AtNs:     int64(r.At),
		Kind:     string(r.Kind),
		NodeID:   r.NodeID,
		LinkID:   r.LinkID,
		PacketID: r.PacketID,
		Reason:   string(r.Reason),
		Summary:  r.Summary,
		Details:  r.Details,
	})
}

// UnmarshalJSON decodes a record produced by MarshalJSON.
func (r *Record) UnmarshalJSON(data []byte) error {
	var w wireRecord
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	*r = Record{
		Index:    w.Index,
		At:       sim.SimTime(w.AtNs),
		Kind:     sim.EventKind(w.Kind),
		NodeID:   w.NodeID,
		LinkID:   w.LinkID,
		PacketID: w.PacketID,
		Reason:   sim.ReasonCode(w.Reason),
		Summary:  w.Summary,
		Details:  w.Details,
	}
	return nil
}
