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

func TestCandidateDecisionEvidenceExposesSourceRefs(t *testing.T) {
	root := repoRootForTest(t)
	out := t.TempDir() + "/candidates.json"
	evidenceOut := t.TempDir() + "/evidence.json"
	if err := generateCandidate(t, root, out); err != nil {
		t.Fatal(err)
	}
	if err := run(options{Repo: "../..", Manifest: defaultManifest, CandidateIn: out, EvidenceOut: evidenceOut}); err != nil {
		t.Fatalf("run: %v", err)
	}
	var got evidence
	if err := readJSON(evidenceOut, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.CandidateSourceRefs) != 1 ||
		got.CandidateSourceRefs[0].SourceRef.SourceWorkflow == "" {
		t.Fatalf("candidate source refs = %+v", got.CandidateSourceRefs)
	}
}
