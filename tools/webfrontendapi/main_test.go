package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunWritesEvidence(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	if err := mainRun([]string{"-repo", "../..", "-evidence-out", out}); err != nil {
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
	if got.Status != "verified" || got.CasesVerified != 4 || got.SourceChecks == 0 {
		t.Fatalf("unexpected evidence: %+v", got)
	}
}

func TestGoRunWiresCLIFlags(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	cmd := exec.Command("go", "run", ".", "-repo", "../..", "-evidence-out", out)
	if body, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go run webfrontendapi: %v\n%s", err, body)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("missing evidence from go run: %v", err)
	}
}

func TestGeneratedDocMatchesManifest(t *testing.T) {
	if err := mainRun([]string{"-repo", "../..", "-check-doc"}); err != nil {
		t.Fatal(err)
	}
}
