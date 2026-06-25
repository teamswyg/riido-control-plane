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

func TestCandidateIntakeEvidenceExposesSourceRefs(t *testing.T) {
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
	if len(got.CandidateSourceRefs) != 1 ||
		got.CandidateSourceRefs[0].SourceRef.Run.ID == "" {
		t.Fatalf("candidate source refs = %+v", got.CandidateSourceRefs)
	}
}
