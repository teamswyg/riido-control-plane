package main

import "testing"

func TestCandidateDecisionRefreshCommandsCarryCandidateSubject(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	writeFirstCandidateSubject(
		t,
		out,
		`{"kind":"claim_coverage_gap","claim_id":"claim_one"}`,
	)
	result, err := verifyCandidateDecisions(
		root,
		manifestWithGeneratedCandidateRecord(t),
		out,
	)
	if err != nil {
		t.Fatalf("verify candidate decision: %v", err)
	}
	got := newRefreshCommandEvidence(result)
	if got.CommandCount != 1 {
		t.Fatalf("command count = %d", got.CommandCount)
	}
	if got.Commands[0].CandidateID != result.DecisionArtifacts[0].CandidateID {
		t.Fatalf("candidate id = %+v", got.Commands[0])
	}
	if got.Commands[0].SubjectKind != "claim_coverage_gap" {
		t.Fatalf("subject kind = %+v", got.Commands[0])
	}
}
