package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestModuleDecompositionEvidence(t *testing.T) {
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
	if got.PackageCount < 20 || got.ToolPackages == 0 || got.ForbiddenImportHit != 0 {
		t.Fatalf("unexpected evidence: %+v", got)
	}
}

func TestModuleDecompositionGeneratedDocFresh(t *testing.T) {
	if err := mainRun([]string{"-repo", "../..", "-check-doc"}); err != nil {
		t.Fatalf("check doc: %v", err)
	}
}
