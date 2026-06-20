package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesEvidence(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := mainRun([]string{"-evidence-out", out}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != evidenceSchema || got.Routes != len(requiredRoutes) {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if got.Loop.Observation == "" || got.Loop.Retrospective == "" {
		t.Fatalf("missing loop evidence: %+v", got.Loop)
	}
}

func TestGeneratedDocMatchesManifest(t *testing.T) {
	if err := mainRun([]string{"-check-doc"}); err != nil {
		t.Fatal(err)
	}
}
