package main

import "testing"

func TestClosedLoopCandidateIntakeManifestVerifies(t *testing.T) {
	if err := run(options{
		Repo:        "../..",
		Manifest:    defaultManifest,
		CheckDoc:    true,
		EvidenceOut: t.TempDir() + "/evidence.json",
	}); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestClosedLoopCandidateIntakeVerifyAlias(t *testing.T) {
	if err := mainRun([]string{
		"-repo", "../..",
		"-manifest", defaultManifest,
		"-verify",
		"-evidence-out", t.TempDir() + "/evidence.json",
	}); err != nil {
		t.Fatalf("mainRun -verify: %v", err)
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
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	return m
}
