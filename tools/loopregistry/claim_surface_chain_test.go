package main

import "testing"

func TestClaimEvidenceChainCoverageRejectsMissingClaim(t *testing.T) {
	claim := claimBinding{ID: "uncovered_claim"}
	if err := verifyClaimEvidenceChains([]claimBinding{claim}, map[string][]string{}); err == nil {
		t.Fatal("expected missing evidence graph chain to fail")
	}
}

func TestClaimSurfaceEvidenceIncludesEvidenceGraphChainIDs(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll(root, m, hashes)
	if err != nil {
		t.Fatalf("verifyAll: %v", err)
	}
	for _, surface := range result.ClaimSurfaces {
		if len(surface.EvidenceChainIDs) == 0 {
			t.Fatalf("claim surface %s missing evidence chain ids", surface.ID)
		}
	}
}
