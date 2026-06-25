package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCandidateIntakeEvidenceOutputExposesAdoptionPlan(t *testing.T) {
	root := repoRootForTest(t)
	candidateIn := candidateFixtureForTest(t, root)
	evidenceOut := t.TempDir() + "/evidence.json"
	if err := run(options{
		Repo:        "../..",
		Manifest:    defaultManifest,
		CandidateIn: candidateIn,
		EvidenceOut: evidenceOut,
	}); err != nil {
		t.Fatalf("run: %v", err)
	}
	data, err := os.ReadFile(evidenceOut)
	if err != nil {
		t.Fatal(err)
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.CandidateAdoption) != got.CandidateCount {
		t.Fatalf("expected adoption plan evidence for each candidate, got %d/%d",
			len(got.CandidateAdoption), got.CandidateCount)
	}
}
