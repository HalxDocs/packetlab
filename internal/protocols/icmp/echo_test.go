package icmp

import (
	"bytes"
	"errors"
	"testing"
)

func TestEchoRoundTrip(t *testing.T) {
	in := Request(7, 42, []byte("packetlab"))
	encoded := Encode(in)
	out, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if out.Type != TypeEchoRequest {
		t.Fatalf("Type = %d, want %d", out.Type, TypeEchoRequest)
	}
	if out.Echo.ID != 7 || out.Echo.Seq != 42 {
		t.Fatalf("echo = %+v", out.Echo)
	}
	if !bytes.Equal(out.Echo.Data, []byte("packetlab")) {
		t.Fatalf("data = %q", out.Echo.Data)
	}
}

func TestEchoWithoutData(t *testing.T) {
	out, err := Decode(Encode(Request(1, 2, nil)))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(out.Echo.Data) != 0 {
		t.Fatalf("data = %v, want empty", out.Echo.Data)
	}
}

func TestReplyTo(t *testing.T) {
	request := Request(9, 3, []byte("abc"))
	reply := ReplyTo(request)
	if !reply.IsReply() {
		t.Fatal("ReplyTo did not produce a reply")
	}
	if reply.Echo.ID != 9 || reply.Echo.Seq != 3 || string(reply.Echo.Data) != "abc" {
		t.Fatalf("reply = %+v", reply.Echo)
	}
	if request.IsReply() {
		t.Fatal("request should not be a reply")
	}
}

func TestDecodeMalformed(t *testing.T) {
	cases := [][]byte{
		nil,
		{8},
		{8, 0, 1, 0},
		{99, 0, 1, 0, 1},
	}
	for _, b := range cases {
		if _, err := Decode(b); err == nil {
			t.Errorf("Decode(%v) succeeded, want error", b)
		}
	}
}

func TestDecodeErrorIsSentinel(t *testing.T) {
	if _, err := Decode([]byte{1}); !errors.Is(err, ErrMalformed) {
		t.Fatalf("err = %v, want ErrMalformed", err)
	}
}

func TestDecodeCopiesData(t *testing.T) {
	encoded := Encode(Request(1, 1, []byte("xy")))
	out, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	encoded[headerBytes] = 'z'
	if out.Echo.Data[0] != 'x' {
		t.Fatal("decoded data aliases the input buffer")
	}
}
