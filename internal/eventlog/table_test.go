package eventlog

import (
	"bytes"
	"strings"
	"testing"

	"github.com/HalxDocs/packetlab/internal/sim"
)

func TestWriteTable(t *testing.T) {
	records := []Record{
		{Index: 0, At: 0, Kind: sim.PacketCreated, NodeID: "host-a", PacketID: "p-001", Summary: "created ICMP echo request"},
		{Index: 1, At: sim.Ms(1), Kind: sim.PacketDropped, NodeID: "router-r1", PacketID: "p-001", Reason: sim.ReasonLinkDown, Summary: "dropped: link is down"},
	}
	var buf bytes.Buffer
	if err := WriteTable(&buf, records); err != nil {
		t.Fatalf("WriteTable: %v", err)
	}
	out := buf.String()

	for _, want := range []string{"TIME", "KIND", "NODE", "PACKET", "SUMMARY", "0.000ms", "1.000ms", "packet.created", "packet.dropped"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want header plus 2 rows:\n%s", len(lines), out)
	}
}

func TestWriteTableUsesDashForMissingIDs(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTable(&buf, []Record{{Kind: sim.RunStarted, Summary: "simulation started"}}); err != nil {
		t.Fatalf("WriteTable: %v", err)
	}
	if !strings.Contains(buf.String(), "-") {
		t.Fatalf("expected placeholder dashes:\n%s", buf.String())
	}
}
