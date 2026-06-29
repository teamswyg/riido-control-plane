package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCandidateIntakeRunEvidencePreservesCandidateSubject(t *testing.T) {
	root := repoRootForTest(t)
	candidateIn := candidateFixtureForTest(t, root)
	writeFirstCandidateSubject(t, candidateIn, `{"kind":"claim_coverage_gap","claim_id":"claim_one"}`)
	evidenceOut := t.TempDir() + "/evidence.json"
	if err := run(options{
		Repo:        "../..",
		Manifest:    defaultManifest,
		CandidateIn: candidateIn,
		EvidenceOut: evidenceOut,
	}); err != nil {
		t.Fatal(err)
	}
	var got evidence
	data, err := os.ReadFile(evidenceOut)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.CandidateSubjects) != 1 {
		t.Fatalf("evidence subjects = %+v", got.CandidateSubjects)
	}
}
