package eventlog

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// WriteTable writes a human-readable trace table.
func WriteTable(w io.Writer, records []Record) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "TIME\tKIND\tNODE\tPACKET\tSUMMARY"); err != nil {
		return err
	}
	for _, rec := range records {
		node := rec.NodeID
		if node == "" {
			node = "-"
		}
		packet := rec.PacketID
		if packet == "" {
			packet = "-"
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			rec.At.TraceString(), rec.Kind, node, packet, rec.Summary); err != nil {
			return err
		}
	}
	return tw.Flush()
}
