package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot-gate.json")
	if err := run([]string{"-check-doc", "-evidence-out", path}); err != nil {
		t.Fatalf("run: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("decode evidence: %v", err)
	}
	if got.SchemaVersion != evidenceSchema || got.Status != "verified" {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if got.OperationsVerified < len(requiredOperations) || got.SignalsVerified < len(requiredSignals) {
		t.Fatalf("missing evidence counts: %+v", got)
	}
	if got.Loop.Observation == "" || got.Loop.Retrospective == "" {
		t.Fatalf("missing loop evidence: %+v", got.Loop)
	}
}
