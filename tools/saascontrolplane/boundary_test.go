package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBoundaryEvidence(t *testing.T) {
	out := filepath.Join(t.TempDir(), "evidence.json")
	err := mainRun([]string{"-repo", "../..", "-boundary", "assignment-contract", "-evidence-out", out})
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.BoundaryID != "assignment-contract" || got.BoundaryArtifact != "assignment-contract-evidence" {
		t.Fatalf("unexpected boundary evidence: %+v", got)
	}
}

func TestUnknownBoundaryFails(t *testing.T) {
	err := mainRun([]string{"-repo", "../..", "-boundary", "missing"})
	if err == nil {
		t.Fatal("expected unknown boundary to fail")
	}
}
