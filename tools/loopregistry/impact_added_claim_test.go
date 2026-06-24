package main

import "testing"

func TestAddedClaimImpactRequiresBoundFileChange(t *testing.T) {
	current := []claimBinding{{ID: "claim", Statement: "new", Loop: "loop", Files: []string{"a.go"}}}
	if _, err := verifyClaimImpact("origin/main", nil, current,
		map[string]bool{defaultManifest: true, evidenceGraphManifest: true}); err == nil {
		t.Fatal("expected added claim without bound file change to fail")
	}
}

func TestAddedClaimImpactAllowsBoundFileAndReasoningChange(t *testing.T) {
	current := []claimBinding{{ID: "claim", Statement: "new", Loop: "loop", Files: []string{"a.go"}}}
	evidence, err := verifyClaimImpact("origin/main", nil, current,
		map[string]bool{"a.go": true, defaultManifest: true, evidenceGraphManifest: true})
	if err != nil {
		t.Fatalf("verify impact: %v", err)
	}
	if evidence.AddedClaimCount != 1 || len(evidence.AddedClaims[0].ChangedReasoningEvidence) != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}
