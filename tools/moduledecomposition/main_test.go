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
	if got.LineBudgetTarget == 0 || got.FilesOverLineBudget == 0 || len(got.LineBudgetSamples) == 0 {
		t.Fatalf("missing line budget evidence: %+v", got)
	}
	if len(got.LineBudgetHotspots) == 0 {
		t.Fatalf("missing line budget hotspot evidence: %+v", got)
	}
	if got.LineBudgetRatchet.MaxFilesOverTarget == 0 || got.LineBudgetRatchet.MaxFileLines == 0 {
		t.Fatalf("missing line budget ratchet evidence: %+v", got)
	}
	if len(got.LineBudgetHotspotRatchets) == 0 {
		t.Fatalf("missing line budget hotspot ratchet evidence: %+v", got)
	}
}

func TestModuleDecompositionGeneratedDocFresh(t *testing.T) {
	if err := mainRun([]string{"-repo", "../..", "-check-doc"}); err != nil {
		t.Fatalf("check doc: %v", err)
	}
}

func TestLineBudgetRatchetRejectsWorseEvidence(t *testing.T) {
	err := verifyLineBudgetRatchet(lineBudgetResult{
		OverTarget:         2,
		MaxLines:           101,
		MaxFilesOverTarget: 1,
		MaxFileLinesLimit:  100,
	})
	if err == nil {
		t.Fatal("expected ratchet failure")
	}
}

func TestLineBudgetHotspotRatchetRejectsWorseEvidence(t *testing.T) {
	err := verifyLineBudgetHotspotRatchets([]lineBudgetHotspotRatchet{
		{Path: "internal/example", FilesSlack: -1},
	})
	if err == nil {
		t.Fatal("expected hotspot ratchet failure")
	}
}
