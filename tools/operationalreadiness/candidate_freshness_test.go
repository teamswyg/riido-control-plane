package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestOperationalReadinessCandidateArtifactIsFresh(t *testing.T) {
	dir := t.TempDir()
	out := dir + "/candidates.json"
	t.Setenv(readinessNowEnv, "2026-06-29T12:00:00Z")
	if err := run(options{Repo: "../..", Manifest: defaultManifest, CandidateOut: out}); err != nil {
		t.Fatal(err)
	}
	got := readCandidateEvidence(t, out)
	if got.CandidateCount == 0 || got.SourceExpiresAt == "" {
		t.Fatalf("candidate artifact = %+v", got)
	}
	if got.SourceGeneratedAt != "2026-06-29T12:00:00Z" ||
		got.SourceExpiresAt != "2026-06-30T12:00:00Z" {
		t.Fatalf("candidate freshness = %+v", got)
	}
	if got.Candidates[0].SourceRef.SummaryArtifact != "operational-readiness-evidence" {
		t.Fatalf("candidate source ref = %+v", got.Candidates[0].SourceRef)
	}
	if got.Candidates[0].SourceRef.SourceExpiresAt != got.SourceExpiresAt {
		t.Fatalf("candidate source freshness = %+v", got.Candidates[0].SourceRef)
	}
}

func readCandidateEvidence(t *testing.T, path string) candidateEvidence {
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
