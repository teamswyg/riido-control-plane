package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const workflowEvidenceGoldenSHA256 = "2c5b6192fc4fe586f064a6a5c8284d5a68d2bf2f6df306cdb58e403481e108b2"

func TestWorkflowEvidence(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-24T00:00:00Z")
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := mainRun([]string{"-repo", "../..", "-evidence-out", out}); err != nil {
		t.Fatalf("run workflow evidence: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read evidence bytes: %v", err)
	}
	if got := sha256Hex(data); got != workflowEvidenceGoldenSHA256 {
		t.Fatalf("workflow evidence golden sha = %s, want %s", got, workflowEvidenceGoldenSHA256)
	}
	var got evidence
	if err := readJSON(out, &got); err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if got.Status != "verified" || got.WorkflowCount == 0 || got.CoveredCount == 0 {
		t.Fatalf("unexpected evidence: %+v", got)
	}
}

func readJSON(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}
