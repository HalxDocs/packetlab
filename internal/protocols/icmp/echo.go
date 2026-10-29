// Package icmp implements the small ICMP-like echo slice PacketLab needs.
// The encoding is deliberately simplified: it is not wire-compatible with
// real ICMP because PacketLab does not model checksums or the full header.
package icmp

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Message types used by PacketLab.
const (
	TypeEchoReply   byte = 0
	TypeEchoRequest byte = 8
)

// headerBytes is type (1) + identifier (2) + sequence (2).
const headerBytes = 5

// ErrMalformed is returned when a payload cannot be decoded.
var ErrMalformed = errors.New("icmp: malformed echo payload")

// Echo is the payload of an echo request or reply.
type Echo struct {
	ID   uint16
	Seq  uint16
	Data []byte
}

// Message is a decoded echo message.
type Message struct {
	Type byte
	Echo Echo
}

// IsRequest reports whether the message is an echo request.
func (m Message) IsRequest() bool { return m.Type == TypeEchoRequest }

// IsReply reports whether the message is an echo reply.
func (m Message) IsReply() bool { return m.Type == TypeEchoReply }

// Encode serializes a message: type, identifier, sequence, then data.
func Encode(m Message) []byte {
	buf := make([]byte, headerBytes+len(m.Echo.Data))
	buf[0] = m.Type
	binary.BigEndian.PutUint16(buf[1:3], m.Echo.ID)
	binary.BigEndian.PutUint16(buf[3:5], m.Echo.Seq)
	copy(buf[headerBytes:], m.Echo.Data)
	return buf
}

// Decode parses a message produced by Encode.
func Decode(b []byte) (Message, error) {
	if len(b) < headerBytes {
		return Message{}, fmt.Errorf("%w: need at least %d bytes, got %d", ErrMalformed, headerBytes, len(b))
	}
	if b[0] != TypeEchoRequest && b[0] != TypeEchoReply {
		return Message{}, fmt.Errorf("%w: unknown type %d", ErrMalformed, b[0])
	}
	return Message{
		Type: b[0],
		Echo: Echo{
			ID:   binary.BigEndian.Uint16(b[1:3]),
			Seq:  binary.BigEndian.Uint16(b[3:5]),
			Data: append([]byte(nil), b[headerBytes:]...),
		},
	}, nil
}

// Request builds an echo request message.
func Request(id, seq uint16, data []byte) Message {
	return Message{Type: TypeEchoRequest, Echo: Echo{ID: id, Seq: seq, Data: data}}
}

// ReplyTo builds an echo reply carrying the same identifier, sequence, and
// data as the request.
func ReplyTo(request Message) Message {
	return Message{Type: TypeEchoReply, Echo: request.Echo}
}
