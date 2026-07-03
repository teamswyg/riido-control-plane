package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationLedgerBehaviorGolden(t *testing.T) {
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
	if got.SchemaVersion != evidenceSchema || got.ID != expectedID || got.Status != "verified" {
		t.Fatalf("unexpected evidence: %+v", got)
	}
	if got.SectionCount != 98 || got.SliceCount != 90 || got.ValidationGates != 6 || got.RiidoReferences != 100 {
		t.Fatalf("unexpected evidence counts: %+v", got)
	}
	if got.GeneratedDoc != "docs/migration/control-plane.md" || got.EvidenceArtifact != "migration-ledger-evidence" {
		t.Fatalf("unexpected output refs: %+v", got)
	}
	if got.Loop.Observation != "The control-plane migration history was the last registered manual documentation debt and could not prove section coverage without reading a long prose file." {
		t.Fatalf("unexpected loop observation: %s", got.Loop.Observation)
	}
}

func TestMigrationLedgerGeneratedDocFresh(t *testing.T) {
	if err := mainRun([]string{"-repo", "../..", "-check-doc"}); err != nil {
		t.Fatalf("check doc: %v", err)
	}
}
