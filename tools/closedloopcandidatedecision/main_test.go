package main

import "testing"

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
	root := repoRootForTest(t)
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	return m
}
