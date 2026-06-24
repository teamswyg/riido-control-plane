package main

import (
	"strings"
	"testing"
)

func TestCandidateIntakeRejectsMissingAdoptionPlan(t *testing.T) {
	m := loadIntakeManifestForTest(t)
	item := closedLoopCandidate{
		ID: "smoke:broken", HarnessLoop: "provider_acceptance_harness",
		PromotionTarget: m.Sources[0].PromotionTarget, Observation: "broken",
		RequiredNextArtifacts: m.Sources[0].RequiredNextArtifacts,
	}
	err := verifyCandidateItem(m, item)
	if err == nil || !strings.Contains(err.Error(), "adoption plan") {
		t.Fatalf("expected adoption plan failure, got %v", err)
	}
}
