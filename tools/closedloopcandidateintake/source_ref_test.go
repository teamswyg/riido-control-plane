package main

import "testing"

func TestCandidateIntakeRejectsSourceRefDrift(t *testing.T) {
	root := repoRootForTest(t)
	path := candidateFixtureForTest(t, root)
	candidate, _, err := loadCandidate(path)
	if err != nil {
		t.Fatal(err)
	}
	candidate.Candidates[0].SourceRef.CandidateArtifact = "other-candidates"
	drifted := t.TempDir() + "/drifted-candidates.json"
	if err := writeJSON(drifted, candidate); err != nil {
		t.Fatal(err)
	}
	if _, err := verifyCandidateFile(root, loadIntakeManifestForTest(t), drifted); err == nil {
		t.Fatal("expected source_ref drift to fail")
	}
}

func TestCandidateIntakeRejectsMissingSourceRefRun(t *testing.T) {
	root := repoRootForTest(t)
	path := candidateFixtureForTest(t, root)
	candidate, _, err := loadCandidate(path)
	if err != nil {
		t.Fatal(err)
	}
	candidate.Candidates[0].SourceRef.Run.ID = ""
	drifted := t.TempDir() + "/missing-run-candidates.json"
	if err := writeJSON(drifted, candidate); err != nil {
		t.Fatal(err)
	}
	if _, err := verifyCandidateFile(root, loadIntakeManifestForTest(t), drifted); err == nil {
		t.Fatal("expected missing source_ref run to fail")
	}
}
