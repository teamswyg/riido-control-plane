package main

import (
	"strings"
	"testing"
)

func TestCandidateDecisionRejectsMissingAdoptionPlan(t *testing.T) {
	err := verifyAdoptionPlan(closedLoopCandidate{
		ID:                    "smoke:broken",
		RequiredNextArtifacts: []string{"claim_binding"},
	})
	if err == nil || !strings.Contains(err.Error(), "adoption plan") {
		t.Fatalf("expected adoption plan failure, got %v", err)
	}
}
