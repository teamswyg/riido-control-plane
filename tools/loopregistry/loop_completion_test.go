package main

import "testing"

func TestLoopCompletionEvidenceCoversEveryLoop(t *testing.T) {
	m := loadLoopRegistryManifestForTest(t)
	items := loopCompletions(m)
	summary := summarizeLoopCompletions(items)
	if summary.LoopCount != len(m.Loops) {
		t.Fatalf("loop count mismatch: %+v", summary)
	}
	if summary.BelowThresholdCount != 0 {
		t.Fatalf("unexpected partial loops: %+v", summary)
	}
	if summary.MinCompletionBasisPoints < loopCompletionThresholdBasisPoints {
		t.Fatalf("completion below threshold: %+v", summary)
	}
}

func TestLoopCompletionScoresMissingGraphEvidence(t *testing.T) {
	m := loadLoopRegistryManifestForTest(t)
	m.EvidenceGraph = nil
	summary := summarizeLoopCompletions(loopCompletions(m))
	if summary.MinCompletionBasisPoints >= 10000 {
		t.Fatalf("expected graph gap to reduce completion: %+v", summary)
	}
}

func TestLoopCompletionRejectsMissingClaimCoverage(t *testing.T) {
	m := loadLoopRegistryManifestForTest(t)
	m.Claims = nil
	if err := verifyLoopCompletions(m); err == nil {
		t.Fatal("expected missing claim coverage to fail")
	}
}

func loadLoopRegistryManifestForTest(t *testing.T) manifest {
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
