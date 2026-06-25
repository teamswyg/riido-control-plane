package main

import "testing"

func TestCandidateDecisionRejectsSourceRefDrift(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/candidates.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	candidate, _, err := loadCandidate(out)
	if err != nil {
		t.Fatal(err)
	}
	candidate.Candidates[0].SourceRef.SourceWorkflow = ".github/workflows/other.yml"
	drifted := t.TempDir() + "/drifted-candidates.json"
	if err := writeJSON(drifted, candidate); err != nil {
		t.Fatal(err)
	}
	if _, err := verifyCandidateDecisions(root, loadDecisionManifestForTest(t), drifted); err == nil {
		t.Fatal("expected source_ref drift to fail")
	}
}
