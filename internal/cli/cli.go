// Package cli implements the packetlab command line: run, validate, trace,
// and version. It is a transport adapter over the simulation packages.
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/HalxDocs/packetlab/internal/engine"
	"github.com/HalxDocs/packetlab/internal/eventlog"
	"github.com/HalxDocs/packetlab/internal/scenario"
	"github.com/HalxDocs/packetlab/internal/session"
	"github.com/HalxDocs/packetlab/internal/topology"
	"github.com/HalxDocs/packetlab/internal/trace"
)

// Version is the CLI version. It can be overridden at build time with
// -ldflags "-X github.com/HalxDocs/packetlab/internal/cli.Version=...".
var Version = "0.1.0"

// Exit codes.
const (
	ExitOK    = 0
	ExitError = 1
	ExitUsage = 2
)

// Run dispatches a command line and returns a process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return ExitUsage
	}
	switch args[0] {
	case "run":
		return runCommand(args[1:], stdout, stderr)
	case "validate":
		return validateCommand(args[1:], stdout, stderr)
	case "trace":
		return traceCommand(args[1:], stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "packetlab %s\n", Version)
		return ExitOK
	case "help", "--help", "-h":
		usage(stdout)
		return ExitOK
	default:
		fmt.Fprintf(stderr, "packetlab: unknown command %q\n\n", args[0])
		usage(stderr)
		return ExitUsage
	}
}

func usage(w io.Writer) {
	fmt.Fprint(w, `PacketLab — a deterministic network simulator

Usage:
  packetlab run [--json] [--limit N] <scenario.json>
  packetlab validate <scenario.json>
  packetlab trace [--limit N] <scenario.json>
  packetlab version
  packetlab help

Commands:
  run       Run a scenario and print its event trace
  validate  Validate a scenario without running it
  trace     Emit a JSON bundle for the web viewer
  version   Print the version
`)
}

func runCommand(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "emit one JSON record per line")
	limit := fs.Int("limit", 0, "stop after N events (0 = engine limit)")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: packetlab run [--json] [--limit N] <scenario.json>")
	}
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return ExitUsage
	}

	s, topo, opts, code := loadScenario(fs.Arg(0), stderr)
	if code != ExitOK {
		return code
	}
	sess, err := session.New(s.ID, topo, opts)
	if err != nil {
		fail(stderr, err)
		return ExitError
	}
	if err := sess.Run(*limit); err != nil {
		fail(stderr, err)
		return ExitError
	}

	records := sess.Records()
	if *jsonOut {
		if err := eventlog.EncodeJSONL(stdout, records); err != nil {
			fail(stderr, err)
			return ExitError
		}
		return ExitOK
	}

	fmt.Fprintf(stdout, "scenario %s\n", s.Summary())
	if s.Description != "" {
		fmt.Fprintf(stdout, "%s\n", s.Description)
	}
	fmt.Fprintf(stdout, "topology %d nodes, %d links; %d action(s)\n\n",
		topo.NodeCount(), topo.LinkCount(), len(opts.Actions))

	if err := eventlog.WriteTable(stdout, records); err != nil {
		fail(stderr, err)
		return ExitError
	}

	result := sess.Result()
	fmt.Fprintf(stdout, "\nresult: %s in %s (%d records, %d delivered, %d dropped)\n",
		stateLabel(sess.State()), result.Duration.TraceString(),
		len(records), result.Delivered, result.Dropped)
	return ExitOK
}

func validateCommand(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: packetlab validate <scenario.json>")
	}
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return ExitUsage
	}
	s, topo, opts, code := loadScenario(fs.Arg(0), stderr)
	if code != ExitOK {
		return code
	}
	fmt.Fprintf(stdout, "ok %s: %d nodes, %d links, %d actions\n",
		s.ID, topo.NodeCount(), topo.LinkCount(), len(opts.Actions))
	return ExitOK
}

func traceCommand(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("trace", flag.ContinueOnError)
	fs.SetOutput(stderr)
	limit := fs.Int("limit", 0, "stop after N events (0 = engine limit)")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: packetlab trace [--limit N] <scenario.json>")
	}
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return ExitUsage
	}

	s, topo, opts, code := loadScenario(fs.Arg(0), stderr)
	if code != ExitOK {
		return code
	}
	sess, err := session.New(s.ID, topo, opts)
	if err != nil {
		fail(stderr, err)
		return ExitError
	}
	if err := sess.Run(*limit); err != nil {
		fail(stderr, err)
		return ExitError
	}

	bundle := trace.Build(s, topo, sess.Records(), sess.Result())
	enc := json.NewEncoder(stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(bundle); err != nil {
		fail(stderr, err)
		return ExitError
	}
	return ExitOK
}

// loadScenario loads and builds a scenario, returning a process exit code on
// failure.
func loadScenario(path string, stderr io.Writer) (*scenario.Scenario, *topology.Topology, engine.Options, int) {
	s, err := scenario.Load(path)
	if err != nil {
		fail(stderr, err)
		return nil, nil, engine.Options{}, ExitError
	}
	topo, opts, err := s.Build()
	if err != nil {
		fail(stderr, err)
		return nil, nil, engine.Options{}, ExitError
	}
	return s, topo, opts, ExitOK
}

func fail(w io.Writer, err error) {
	fmt.Fprintf(w, "packetlab: %v\n", err)
}

func stateLabel(state session.State) string {
	switch state {
	case session.StateCompleted:
		return "completed"
	case session.StateLimit:
		return "stopped at event limit"
	case session.StatePaused:
		return "paused"
	case session.StateRunning:
		return "running"
	default:
		return string(state)
	}
}
