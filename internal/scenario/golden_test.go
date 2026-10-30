package scenario

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/HalxDocs/packetlab/internal/engine"
	"github.com/HalxDocs/packetlab/internal/eventlog"
)

// TestGoldenTraces runs canonical scenarios and compares their normalized
// event logs byte for byte. Set UPDATE_GOLDEN=1 to rewrite the files.
func TestGoldenTraces(t *testing.T) {
	files := []string{
		"two-hosts-one-router",
		"missing-route",
		"link-down",
		"ttl-loop",
	}
	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			s, err := Load(scenarioPath(name + ".json"))
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			topo, opts, err := s.Build()
			if err != nil {
				t.Fatalf("Build: %v", err)
			}
			e, err := engine.New(topo, opts)
			if err != nil {
				t.Fatalf("engine.New: %v", err)
			}
			if err := e.RunToCompletion(); err != nil {
				t.Fatalf("Run: %v", err)
			}

			var got bytes.Buffer
			if err := eventlog.EncodeJSONL(&got, e.Records()); err != nil {
				t.Fatalf("EncodeJSONL: %v", err)
			}

			goldenPath := filepath.Join("..", "..", "testdata", "golden", name+".jsonl")
			if os.Getenv("UPDATE_GOLDEN") == "1" {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
					t.Fatalf("MkdirAll: %v", err)
				}
				if err := os.WriteFile(goldenPath, got.Bytes(), 0o644); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("ReadFile: %v (run with UPDATE_GOLDEN=1)", err)
			}
			if !bytes.Equal(got.Bytes(), want) {
				t.Fatalf("golden trace mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got.Bytes(), want)
			}
		})
	}
}
