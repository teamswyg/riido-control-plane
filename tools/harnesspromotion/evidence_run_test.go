package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestHarnessPromotionRunWritesPromotionResult(t *testing.T) {
	t.Setenv("RIIDO_HARNESS_PROMOTION_NOW", "2026-06-24T12:00:00Z")
	dir := t.TempDir()
	evidencePath := dir + "/evidence.json"
	candidatePath := dir + "/candidates.json"
	if err := run(options{
		Repo:         "../..",
		Manifest:     defaultManifest,
		Summary:      "docs/30-architecture/fixtures/performance-failure-summary.fixture.json",
		CandidateOut: candidatePath,
		EvidenceOut:  evidencePath,
	}); err != nil {
		t.Fatalf("run: %v", err)
	}
	got := loadPromotionEvidenceForTest(t, evidencePath)
	if got.PromotionResult == nil {
		t.Fatal("missing promotion result")
	}
	if got.PromotionResult.CandidateArtifact != candidatePath ||
		got.PromotionResult.CandidateCount != 2 ||
		len(got.PromotionResult.CandidateIDs) != 2 {
		t.Fatalf("promotion result = %+v", got.PromotionResult)
	}
}

func loadPromotionEvidenceForTest(t *testing.T, path string) evidence {
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
