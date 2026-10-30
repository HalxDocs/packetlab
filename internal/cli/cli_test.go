package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HalxDocs/packetlab/internal/eventlog"
	"github.com/HalxDocs/packetlab/internal/trace"
)

func runCLI(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func scenarioPath(name string) string {
	return filepath.Join("..", "..", "scenarios", name)
}

func TestVersion(t *testing.T) {
	code, out, _ := runCLI("version")
	if code != ExitOK {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(out, "packetlab") || !strings.Contains(out, Version) {
		t.Fatalf("output = %q", out)
	}
}

func TestHelp(t *testing.T) {
	code, out, _ := runCLI("help")
	if code != ExitOK {
		t.Fatalf("exit = %d", code)
	}
	for _, want := range []string{"Usage:", "run", "validate", "trace"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q", want)
		}
	}
}

func TestNoArgs(t *testing.T) {
	code, _, errOut := runCLI()
	if code != ExitUsage {
		t.Fatalf("exit = %d, want %d", code, ExitUsage)
	}
	if !strings.Contains(errOut, "Usage:") {
		t.Fatalf("stderr = %q", errOut)
	}
}

func TestUnknownCommand(t *testing.T) {
	code, _, errOut := runCLI("frobnicate")
	if code != ExitUsage {
		t.Fatalf("exit = %d, want %d", code, ExitUsage)
	}
	if !strings.Contains(errOut, "unknown command") {
		t.Fatalf("stderr = %q", errOut)
	}
}

func TestRunTable(t *testing.T) {
	code, out, errOut := runCLI("run", scenarioPath("two-hosts-one-router.json"))
	if code != ExitOK {
		t.Fatalf("exit = %d, stderr = %s", code, errOut)
	}
	for _, want := range []string{
		"scenario two-hosts-one-router",
		"packet.created",
		"route.selected",
		"packet.delivered",
		"ping.reply",
		"result: completed in 4.000ms",
		"2 delivered, 0 dropped",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestRunJSON(t *testing.T) {
	code, out, errOut := runCLI("run", "--json", scenarioPath("two-hosts-one-router.json"))
	if code != ExitOK {
		t.Fatalf("exit = %d, stderr = %s", code, errOut)
	}
	records, err := eventlog.DecodeJSONL(strings.NewReader(out))
	if err != nil {
		t.Fatalf("DecodeJSONL: %v", err)
	}
	if len(records) != 25 {
		t.Fatalf("records = %d, want 25", len(records))
	}
	if records[0].Kind != "run.started" {
		t.Fatalf("first kind = %q", records[0].Kind)
	}
}

func TestRunLimit(t *testing.T) {
	code, out, errOut := runCLI("run", "--limit", "1", scenarioPath("two-hosts-one-router.json"))
	if code != ExitOK {
		t.Fatalf("exit = %d, stderr = %s", code, errOut)
	}
	if !strings.Contains(out, "result: paused") {
		t.Fatalf("output = %q", out)
	}
}

func TestRunMissingFile(t *testing.T) {
	code, _, errOut := runCLI("run", scenarioPath("nope.json"))
	if code != ExitError {
		t.Fatalf("exit = %d, want %d", code, ExitError)
	}
	if !strings.Contains(errOut, "packetlab:") {
		t.Fatalf("stderr = %q", errOut)
	}
}

func TestRunUsage(t *testing.T) {
	code, _, _ := runCLI("run")
	if code != ExitUsage {
		t.Fatalf("exit = %d, want %d", code, ExitUsage)
	}
}

func TestValidateOK(t *testing.T) {
	code, out, errOut := runCLI("validate", scenarioPath("two-hosts-one-router.json"))
	if code != ExitOK {
		t.Fatalf("exit = %d, stderr = %s", code, errOut)
	}
	if !strings.Contains(out, "ok two-hosts-one-router") {
		t.Fatalf("output = %q", out)
	}
}

func TestValidateInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`{"schemaVersion":"1.0"}`), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	code, _, errOut := runCLI("validate", path)
	if code != ExitError {
		t.Fatalf("exit = %d, want %d", code, ExitError)
	}
	if !strings.Contains(errOut, "scenario validation failed") {
		t.Fatalf("stderr = %q", errOut)
	}
}

func TestTraceBundle(t *testing.T) {
	code, out, errOut := runCLI("trace", scenarioPath("two-hosts-one-router.json"))
	if code != ExitOK {
		t.Fatalf("exit = %d, stderr = %s", code, errOut)
	}
	var bundle trace.Bundle
	if err := json.Unmarshal([]byte(out), &bundle); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(bundle.Records) != 25 {
		t.Fatalf("records = %d", len(bundle.Records))
	}
	if len(bundle.Topology.Nodes) != 3 || len(bundle.Topology.Links) != 2 {
		t.Fatalf("topology = %+v", bundle.Topology)
	}
	if bundle.Result.Delivered != 2 {
		t.Fatalf("result = %+v", bundle.Result)
	}
}

func TestRunDroppedScenario(t *testing.T) {
	code, out, errOut := runCLI("run", scenarioPath("link-down.json"))
	if code != ExitOK {
		t.Fatalf("exit = %d, stderr = %s", code, errOut)
	}
	if !strings.Contains(out, "dropped: link is down") {
		t.Fatalf("output missing drop explanation:\n%s", out)
	}
	if !strings.Contains(out, "1 dropped") {
		t.Fatalf("output missing drop count:\n%s", out)
	}
}

func TestRunTTLScenario(t *testing.T) {
	code, out, errOut := runCLI("run", scenarioPath("ttl-loop.json"))
	if code != ExitOK {
		t.Fatalf("exit = %d, stderr = %s", code, errOut)
	}
	if !strings.Contains(out, "dropped: TTL expired") {
		t.Fatalf("output missing TTL explanation:\n%s", out)
	}
}

func TestTraceMissingFile(t *testing.T) {
	code, _, _ := runCLI("trace", scenarioPath("nope.json"))
	if code != ExitError {
		t.Fatalf("exit = %d, want %d", code, ExitError)
	}
}
