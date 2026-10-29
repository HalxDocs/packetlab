package eventlog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// EncodeJSONL writes one JSON object per record. Field order is fixed by
// Record.MarshalJSON, so identical runs produce identical bytes.
func EncodeJSONL(w io.Writer, records []Record) error {
	bw := bufio.NewWriter(w)
	enc := json.NewEncoder(bw)
	enc.SetEscapeHTML(false)
	for _, rec := range records {
		if err := enc.Encode(rec); err != nil {
			return fmt.Errorf("eventlog: encode record %d: %w", rec.Index, err)
		}
	}
	return bw.Flush()
}

// DecodeJSONL reads records written by EncodeJSONL.
func DecodeJSONL(r io.Reader) ([]Record, error) {
	dec := json.NewDecoder(r)
	var records []Record
	for {
		var rec Record
		err := dec.Decode(&rec)
		if err == io.EOF {
			return records, nil
		}
		if err != nil {
			return nil, fmt.Errorf("eventlog: decode record: %w", err)
		}
		records = append(records, rec)
	}
}
