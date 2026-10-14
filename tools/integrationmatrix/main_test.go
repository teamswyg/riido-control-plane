package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestIntegrationMatrixBehaviorGolden(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := mainRun([]string{"-repo", "../..", "-evidence-out", out}); err != nil {
		t.Fatalf("mainRun: %v", err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	var got evidence
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode evidence: %v", err)
	}
	if got.SchemaVersion != evidenceSchema || got.ID != "control-plane-integration-matrix" || got.Status != "verified" {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if got.PublicGates != 15 || got.PullRequestGates != 12 || got.OperatorGates != 3 || got.PrivateGates != 7 {
		t.Fatalf("unexpected gate counts: %+v", got)
	}
	if got.WorkflowRefs != 15 || got.CommandRefs != 14 {
		t.Fatalf("unexpected ref counts: %+v", got)
	}
	if got.Workflow != ".github/workflows/integration-matrix.yml" || got.GeneratedDoc != "docs/30-architecture/integration-matrix.md" {
		t.Fatalf("unexpected output refs: %+v", got)
	}
}

func TestIntegrationMatrixGeneratedDocFresh(t *testing.T) {
	if err := mainRun([]string{"-repo", "../..", "-check-doc"}); err != nil {
		t.Fatalf("check doc: %v", err)
	}
}
