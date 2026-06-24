package main

import "testing"

func TestRemovedClaimImpactRequiresBoundFileChange(t *testing.T) {
	base := []claimBinding{{ID: "claim", Statement: "old", Loop: "loop", Files: []string{"a.go"}}}
	if _, err := verifyClaimImpact("origin/main", base, nil,
		map[string]bool{defaultManifest: true, evidenceGraphManifest: true}); err == nil {
		t.Fatal("expected removed claim without bound file change to fail")
	}
}

func TestRemovedClaimImpactAllowsBoundFileAndReasoningChange(t *testing.T) {
	base := []claimBinding{{ID: "claim", Statement: "old", Loop: "loop", Files: []string{"a.go"}}}
	evidence, err := verifyClaimImpact("origin/main", base, nil,
		map[string]bool{"a.go": true, defaultManifest: true, evidenceGraphManifest: true})
	if err != nil {
		t.Fatalf("verify impact: %v", err)
	}
	if evidence.RemovedClaimCount != 1 || len(evidence.RemovedClaims[0].ChangedReasoningEvidence) != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}
