package main

import (
	"errors"
	"testing"
)

func TestClosedLoopCandidateIntakeManifestVerifies(t *testing.T) {
	root := repoRootForTest(t)
	candidateIn := candidateFixtureForTest(t, root)
	if err := run(options{
		Repo:        "../..",
		Manifest:    defaultManifest,
		CandidateIn: candidateIn,
		CheckDoc:    true,
		EvidenceOut: t.TempDir() + "/evidence.json",
	}); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestClosedLoopCandidateIntakeVerifyAlias(t *testing.T) {
	root := repoRootForTest(t)
	candidateIn := candidateFixtureForTest(t, root)
	if err := mainRun([]string{
		"-repo", "../..",
		"-manifest", defaultManifest,
		"-candidate-in", candidateIn,
		"-verify",
		"-evidence-out", t.TempDir() + "/evidence.json",
	}); err != nil {
		t.Fatalf("mainRun -verify: %v", err)
	}
}

func TestClosedLoopCandidateIntakeRequiresCandidateInput(t *testing.T) {
	err := run(options{
		Repo:        "../..",
		Manifest:    defaultManifest,
		CheckDoc:    true,
		EvidenceOut: t.TempDir() + "/evidence.json",
	})
	if !errors.Is(err, errMissingCandidateInput) {
		t.Fatalf("expected missing candidate input, got %v", err)
	}
}

func TestCandidateIntakeRejectsMissingGraph(t *testing.T) {
	m := loadIntakeManifestForTest(t)
	m.Sources[0].EvidenceGraphManifest = "missing.json"
	if _, err := verifyAll("../..", m); err == nil {
		t.Fatal("expected missing evidence graph to fail")
	}
}

func TestCandidateIntakeRejectsMissingAdoptionArtifact(t *testing.T) {
	m := loadIntakeManifestForTest(t)
	m.Sources[0].RequiredNextArtifacts = []string{"claim_binding", "verifier", "ci_gate"}
	if _, err := verifyAll("../..", m); err == nil {
		t.Fatal("expected missing required adoption artifact to fail")
	}
}

func loadIntakeManifestForTest(t *testing.T) manifest {
	t.Helper()
	root := repoRootForTest(t)
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	return m
}
