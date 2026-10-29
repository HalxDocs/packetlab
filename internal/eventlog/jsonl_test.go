package eventlog

import (
	"bytes"
	"strings"
	"testing"

	"github.com/HalxDocs/packetlab/internal/sim"
)

func testRecords() []Record {
	return []Record{
		{Index: 0, At: 0, Kind: sim.RunStarted, Summary: "simulation started"},
		{Index: 1, At: sim.Ms(1), Kind: sim.PacketCreated, NodeID: "host-a", PacketID: "p-001", Summary: "created ICMP echo request"},
	}
}

func TestJSONLRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	if err := EncodeJSONL(&buf, testRecords()); err != nil {
		t.Fatalf("EncodeJSONL: %v", err)
	}
	got, err := DecodeJSONL(&buf)
	if err != nil {
		t.Fatalf("DecodeJSONL: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("decoded %d records, want 2", len(got))
	}
	if got[1].PacketID != "p-001" || got[1].At != sim.Ms(1) {
		t.Fatalf("decoded record = %+v", got[1])
	}
}

func TestJSONLIsDeterministic(t *testing.T) {
	var first, second bytes.Buffer
	if err := EncodeJSONL(&first, testRecords()); err != nil {
		t.Fatalf("EncodeJSONL: %v", err)
	}
	if err := EncodeJSONL(&second, testRecords()); err != nil {
		t.Fatalf("EncodeJSONL: %v", err)
	}
	if !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("encoding the same records twice produced different bytes")
	}
}

func TestJSONLOneObjectPerLine(t *testing.T) {
	var buf bytes.Buffer
	if err := EncodeJSONL(&buf, testRecords()); err != nil {
		t.Fatalf("EncodeJSONL: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	for _, line := range lines {
		if !strings.HasPrefix(line, "{") || !strings.HasSuffix(line, "}") {
			t.Fatalf("line is not a JSON object: %q", line)
		}
	}
}

func TestDecodeJSONLEmpty(t *testing.T) {
	got, err := DecodeJSONL(strings.NewReader(""))
	if err != nil {
		t.Fatalf("DecodeJSONL: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("decoded %d records from empty input", len(got))
	}
}

func TestDecodeJSONLRejectsGarbage(t *testing.T) {
	if _, err := DecodeJSONL(strings.NewReader("not json\n")); err == nil {
		t.Fatal("expected decode error")
	}
}
