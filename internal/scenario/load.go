package scenario

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// MaxBytes bounds scenario file size. Imported files are untrusted input.
const MaxBytes = 1 << 20

// Load reads, decodes, and validates a scenario file.
func Load(path string) (*Scenario, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("scenario: open %s: %w", path, err)
	}
	defer f.Close()
	return Decode(f)
}

// Decode reads a scenario from r, rejecting unknown fields and trailing data.
func Decode(r io.Reader) (*Scenario, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("scenario: read: %w", err)
	}
	if len(data) > MaxBytes {
		return nil, fmt.Errorf("scenario: file exceeds %d bytes", MaxBytes)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var s Scenario
	if err := dec.Decode(&s); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("scenario: file is empty")
		}
		return nil, fmt.Errorf("scenario: decode: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("scenario: unexpected data after the scenario object")
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return &s, nil
}

// Marshal encodes a scenario with stable indentation. It is used by tests and
// tooling that writes scenarios back to disk.
func Marshal(s *Scenario) ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}
