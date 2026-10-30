package scenario

import (
	"path/filepath"
	"strings"
	"testing"
)

func scenarioPath(name string) string {
	return filepath.Join("..", "..", "scenarios", name)
}

func TestLoadCanonicalScenarios(t *testing.T) {
	files := []string{
		"two-hosts-one-router.json",
		"missing-route.json",
		"link-down.json",
		"ttl-loop.json",
	}
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			s, err := Load(scenarioPath(file))
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if s.SchemaVersion != SchemaVersion {
				t.Errorf("schema = %q", s.SchemaVersion)
			}
			if s.ID == "" || s.Name == "" {
				t.Errorf("scenario is missing identity: %+v", s)
			}
			if len(s.Nodes) == 0 || len(s.Links) == 0 {
				t.Errorf("scenario is empty: %+v", s)
			}
		})
	}
}

func validJSON() string {
	return `{
		"schemaVersion": "1.0",
		"id": "test",
		"name": "Test",
		"nodes": [
			{
				"id": "a",
				"kind": "host",
				"interfaces": [{"id": "a:eth0", "address": "10.0.0.1/24"}]
			},
			{
				"id": "b",
				"kind": "host",
				"interfaces": [{"id": "b:eth0", "address": "10.0.0.2/24"}]
			}
		],
		"links": [{"id": "l", "a": "a:eth0", "b": "b:eth0", "delayMs": 1}],
		"actions": [{"atMs": 0, "kind": "send_ping", "from": "a", "to": "10.0.0.2"}]
	}`
}

func TestDecodeValid(t *testing.T) {
	s, err := Decode(strings.NewReader(validJSON()))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if s.ID != "test" {
		t.Fatalf("ID = %q", s.ID)
	}
}

func TestDecodeRejectsUnknownField(t *testing.T) {
	bad := strings.Replace(validJSON(), `"id": "test",`, `"id": "test", "surprise": true,`, 1)
	if _, err := Decode(strings.NewReader(bad)); err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestDecodeRejectsTrailingData(t *testing.T) {
	bad := validJSON() + `{"another": true}`
	if _, err := Decode(strings.NewReader(bad)); err == nil {
		t.Fatal("expected trailing data error")
	}
}

func TestDecodeRejectsOversizedInput(t *testing.T) {
	big := strings.Repeat(" ", MaxBytes+1)
	if _, err := Decode(strings.NewReader(big)); err == nil {
		t.Fatal("expected size limit error")
	}
}

func TestDecodeRejectsEmptyInput(t *testing.T) {
	if _, err := Decode(strings.NewReader("")); err == nil {
		t.Fatal("expected empty input error")
	}
}
