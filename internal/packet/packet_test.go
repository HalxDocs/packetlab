package packet

import (
	"testing"

	"github.com/HalxDocs/packetlab/internal/network"
	"github.com/HalxDocs/packetlab/internal/sim"
)

func testPacket(ttl uint8) Packet {
	return Packet{
		ID:        "p-001",
		Src:       network.MustParseIPv4("10.0.1.10"),
		Dst:       network.MustParseIPv4("10.0.2.20"),
		Protocol:  ProtocolICMP,
		TTL:       ttl,
		Payload:   []byte{8, 0, 1, 0, 1},
		CreatedAt: sim.Ms(1),
	}
}

func TestPacketCloneIsDeep(t *testing.T) {
	p := testPacket(64)
	cp := p.Clone()
	cp.Payload[0] = 0

	if p.Payload[0] != 8 {
		t.Fatal("Clone shares the payload slice")
	}
	if cp.ID != p.ID || cp.TTL != p.TTL || !cp.Dst.Equal(p.Dst) {
		t.Fatal("Clone did not preserve packet fields")
	}
}

func TestDecrementTTL(t *testing.T) {
	p := testPacket(64)
	next, ok := p.DecrementTTL()
	if !ok {
		t.Fatal("expected a usable hop")
	}
	if next.TTL != 63 {
		t.Fatalf("TTL = %d, want 63", next.TTL)
	}
	if p.TTL != 64 {
		t.Fatal("DecrementTTL mutated the original packet")
	}
}

func TestDecrementTTLExpires(t *testing.T) {
	for _, ttl := range []uint8{0, 1} {
		if _, ok := testPacket(ttl).DecrementTTL(); ok {
			t.Fatalf("TTL %d should expire", ttl)
		}
	}
}

func TestSizeBytes(t *testing.T) {
	p := testPacket(64)
	if got := p.SizeBytes(); got != 20+8+5 {
		t.Fatalf("SizeBytes = %d, want 33", got)
	}
}

func TestProtocolString(t *testing.T) {
	cases := map[Protocol]string{
		ProtocolICMP: "icmp",
		ProtocolTCP:  "tcp",
		ProtocolUDP:  "udp",
		99:           "unknown",
	}
	for proto, want := range cases {
		if got := proto.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", proto, got, want)
		}
	}
}
