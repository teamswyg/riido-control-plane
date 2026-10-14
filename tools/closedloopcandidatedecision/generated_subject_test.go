package main

import "testing"

func TestCandidateDecisionCommandsCarryGeneratedHarnessSubject(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	result, err := verifyCandidateDecisions(
		root,
		loadDecisionManifestForTest(t),
		out,
	)
	if err != nil {
		t.Fatal(err)
	}
	got := newRefreshCommandEvidence(result)
	if got.CommandCount != 1 {
		t.Fatalf("command count = %d", got.CommandCount)
	}
	if got.Commands[0].SubjectKind != "harness_failure" {
		t.Fatalf("command = %+v", got.Commands[0])
	}
}
