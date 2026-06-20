package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "evidence.json")
	if err := run([]string{"-contract", "../../dependency_allowlist.riido.json", "-evidence-out", path}); err != nil {
		t.Fatalf("run: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	var got evidence
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode evidence: %v", err)
	}
	if got.SchemaVersion != evidenceSchemaVersion || got.Status != "verified" {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if got.DirectDependenciesVerified == 0 || got.AllowedDirectModules == 0 {
		t.Fatalf("missing evidence counts: %+v", got)
	}
}
