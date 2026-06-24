package main

import "testing"

func TestCandidateIntakeRejectsPromotionEdgeDrift(t *testing.T) {
	root := repoRootForTest(t)
	path := candidateFixtureForTest(t, root)
	candidate, _, err := loadCandidate(path)
	if err != nil {
		t.Fatal(err)
	}
	candidate.Candidates[0].PromotionEdge.To = "open_decision_queue"
	drifted := t.TempDir() + "/drifted-candidates.json"
	if err := writeJSON(drifted, candidate); err != nil {
		t.Fatal(err)
	}
	if _, err := verifyCandidateFile(root, loadIntakeManifestForTest(t), drifted); err == nil {
		t.Fatal("expected promotion edge drift to fail")
	}
}

func TestCandidateIntakeEvidenceExposesPromotionEdges(t *testing.T) {
	root := repoRootForTest(t)
	candidateIn := candidateFixtureForTest(t, root)
	evidenceOut := t.TempDir() + "/evidence.json"
	if err := run(options{
		Repo: "../..", Manifest: defaultManifest, CandidateIn: candidateIn, EvidenceOut: evidenceOut,
	}); err != nil {
		t.Fatalf("run: %v", err)
	}
	var got evidence
	if err := readJSON(evidenceOut, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.CandidateEdges) != 1 || got.CandidateEdges[0].Relation != "promotes_failure_to" {
		t.Fatalf("candidate edges = %+v", got.CandidateEdges)
	}
}
