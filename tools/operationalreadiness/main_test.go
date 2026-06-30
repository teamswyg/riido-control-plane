package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestOperationalReadinessVerifies(t *testing.T) {
	out := t.TempDir() + "/evidence.json"
	err := run(options{Repo: "../..", Manifest: defaultManifest, CheckDoc: true, EvidenceOut: out})
	if err != nil {
		t.Fatal(err)
	}
	got := readEvidence(t, out)
	if got.CheckCount < 10 || got.PartialCount == 0 || got.CoveredCount == 0 {
		t.Fatalf("unexpected readiness counts: %+v", got)
	}
	if got.GeneratedAt == "" || got.ExpiresAt == "" || got.GeneratedAt >= got.ExpiresAt {
		t.Fatalf("readiness evidence freshness window missing: %+v", got)
	}
	if got.MeasurementCount < got.CheckCount {
		t.Fatalf("measurement count should cover checks: %+v", got)
	}
	if len(got.MissingCategories) != 0 {
		t.Fatalf("missing categories: %+v", got.MissingCategories)
	}
}

func TestOperationalReadinessRejectsPartialWithoutNextWork(t *testing.T) {
	m := loadManifestForTest(t)
	m.Checks[0].Status = "partial"
	m.Checks[0].NextArtifact = ""
	if err := verifyChecks("../..", m); err == nil {
		t.Fatal("expected partial check without next artifact to fail")
	}
}

func TestOperationalReadinessRejectsMissingLocalEvidence(t *testing.T) {
	m := loadManifestForTest(t)
	m.Checks[0].EvidenceRefs = []evidenceRef{{Kind: "test", Path: "missing/local/evidence.go"}}
	if err := verifyChecks("../..", m); err == nil {
		t.Fatal("expected missing local evidence to fail")
	}
}

func loadManifestForTest(t *testing.T) manifest {
	t.Helper()
	var m manifest
	if err := readJSON("../../"+defaultManifest, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func readEvidence(t *testing.T, path string) evidence {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got
}
