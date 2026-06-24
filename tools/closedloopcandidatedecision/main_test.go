package main

import "testing"

func TestCandidateDecisionManifestVerifies(t *testing.T) {
	if err := run(options{
		Repo:        "../..",
		Manifest:    defaultManifest,
		CheckDoc:    true,
		EvidenceOut: t.TempDir() + "/evidence.json",
	}); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestCandidateDecisionVerifyAlias(t *testing.T) {
	if err := mainRun([]string{
		"-repo", "../..",
		"-manifest", defaultManifest,
		"-verify",
		"-evidence-out", t.TempDir() + "/evidence.json",
	}); err != nil {
		t.Fatalf("mainRun -verify: %v", err)
	}
}

func TestCandidateDecisionRejectsInvalidPriority(t *testing.T) {
	m := loadDecisionManifestForTest(t)
	m.Decisions[0].Priority = "urgent"
	if _, err := verifyAll("../..", m); err == nil {
		t.Fatal("expected invalid priority to fail")
	}
}

func TestCandidateDecisionRejectsUnknownNextLoop(t *testing.T) {
	m := loadDecisionManifestForTest(t)
	m.Decisions[0].NextLoop = "missing_loop"
	if _, err := verifyAll("../..", m); err == nil {
		t.Fatal("expected unknown next loop to fail")
	}
}

func TestCandidateDecisionRejectsMissingReviewBy(t *testing.T) {
	m := loadDecisionManifestForTest(t)
	m.Decisions[0].ReviewBy = ""
	if _, err := verifyAll("../..", m); err == nil {
		t.Fatal("expected missing review_by to fail")
	}
}

func TestCandidateDecisionRejectsExpiredReviewBy(t *testing.T) {
	m := loadDecisionManifestForTest(t)
	m.Decisions[0].ReviewBy = "2000-01-01"
	if _, err := verifyAll("../..", m); err == nil {
		t.Fatal("expected expired review_by to fail")
	}
}

func loadDecisionManifestForTest(t *testing.T) manifest {
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
