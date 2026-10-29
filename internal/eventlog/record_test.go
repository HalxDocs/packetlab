package eventlog

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/HalxDocs/packetlab/internal/sim"
)

func sampleRecord() Record {
	return Record{
		Index:    3,
		At:       sim.Ms(2),
		Kind:     sim.PacketDropped,
		NodeID:   "router-r1",
		LinkID:   "link-b",
		PacketID: "p-001",
		Reason:   sim.ReasonLinkDown,
		Summary:  "dropped: link is down",
		Details:  map[string]string{"ingress": "router-r1:eth0"},
	}
}

func TestRecordJSONFieldOrder(t *testing.T) {
	data, err := json.Marshal(sampleRecord())
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := `{"index":3,"atMs":2,"atNs":2000000,"kind":"packet.dropped","nodeId":"router-r1","linkId":"link-b","packetId":"p-001","reason":"link_down","summary":"dropped: link is down","details":{"ingress":"router-r1:eth0"}}`
	if string(data) != want {
		t.Fatalf("JSON =\n%s\nwant\n%s", data, want)
	}
}

func TestRecordJSONOmitsEmptyFields(t *testing.T) {
	data, err := json.Marshal(Record{Index: 0, Kind: sim.RunStarted, Summary: "simulation started"})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	for _, field := range []string{"nodeId", "linkId", "packetId", "reason", "details"} {
		if strings.Contains(string(data), `"`+field+`"`) {
			t.Errorf("expected %q to be omitted: %s", field, data)
		}
	}
}

func TestRecordJSONRoundTrip(t *testing.T) {
	in := sampleRecord()
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out Record
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Index != in.Index || out.At != in.At || out.Kind != in.Kind {
		t.Fatalf("round trip mismatch: %+v vs %+v", out, in)
	}
	if out.Details["ingress"] != "router-r1:eth0" {
		t.Fatalf("details lost: %v", out.Details)
	}
}
