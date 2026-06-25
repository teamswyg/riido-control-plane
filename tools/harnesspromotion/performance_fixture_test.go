package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestPerformanceFixturePromotesExpectedCandidates(t *testing.T) {
	t.Setenv("RIIDO_HARNESS_PROMOTION_NOW", "2026-06-24T12:00:00Z")
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir() + "/performance-candidates.json"
	err = promoteSummary(root, loadPromotionManifestForTest(t),
		"docs/30-architecture/fixtures/performance-failure-summary.fixture.json", out)
	if err != nil {
		t.Fatalf("promote performance fixture: %v", err)
	}
	got := loadCandidateEvidenceForTest(t, out)
	want := map[string]bool{
		"control-plane-performance:control_plane_performance_hot_path_benchmarks":   true,
		"control-plane-performance:control_plane_performance_closed_loop_promotion": true,
	}
	for _, item := range got.Candidates {
		delete(want, item.ID)
	}
	if len(want) != 0 || got.CandidateCount != 2 {
		t.Fatalf("performance candidates = %+v missing=%+v", got.Candidates, want)
	}
}

func loadCandidateEvidenceForTest(t *testing.T, path string) candidateEvidence {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got candidateEvidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got
}
