package main

import "testing"

func TestClaimImpactRequiresBoundFileChange(t *testing.T) {
	base := []claimBinding{{
		ID:        "claim",
		Statement: "old",
		Loop:      "loop",
		Files:     []string{"internal/example.go"},
	}}
	current := []claimBinding{{
		ID:        "claim",
		Statement: "new",
		Loop:      "loop",
		Files:     []string{"internal/example.go"},
	}}
	if _, err := verifyClaimImpact("origin/main", base, current, map[string]bool{}); err == nil {
		t.Fatal("expected missing bound file change to fail")
	}
}

func TestClaimImpactAllowsBoundFileChange(t *testing.T) {
	base := []claimBinding{{ID: "claim", Statement: "old", Loop: "loop", Files: []string{"a.go"}}}
	current := []claimBinding{{ID: "claim", Statement: "new", Loop: "loop", Files: []string{"a.go"}}}
	evidence, err := verifyClaimImpact("origin/main", base, current, map[string]bool{"a.go": true})
	if err != nil {
		t.Fatalf("verify impact: %v", err)
	}
	if evidence.ChangedClaimCount != 1 || len(evidence.Claims[0].ChangedBoundFiles) != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}
