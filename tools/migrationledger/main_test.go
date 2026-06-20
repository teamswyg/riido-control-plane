package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationLedgerEvidence(t *testing.T) {
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
	if got.Status != "verified" || got.SliceCount < 80 || got.ValidationGates < 5 {
		t.Fatalf("unexpected evidence: %+v", got)
	}
}

func TestMigrationLedgerGeneratedDocFresh(t *testing.T) {
	if err := mainRun([]string{"-repo", "../..", "-check-doc"}); err != nil {
		t.Fatalf("check doc: %v", err)
	}
}
