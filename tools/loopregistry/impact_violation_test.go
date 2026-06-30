package main

import "testing"

func TestClaimImpactReturnsViolationEvidenceOnFailure(t *testing.T) {
	base := []claimBinding{impactClaimFixture("old")}
	current := []claimBinding{impactClaimFixture("new")}
	evidence, err := verifyClaimImpact("origin/main", base, current, map[string]bool{})
	if err == nil {
		t.Fatal("expected impact failure")
	}
	if evidence == nil || len(evidence.Violations) != 1 {
		t.Fatalf("missing violation evidence: %+v", evidence)
	}
	v := evidence.Violations[0]
	if v.Scope != "changed_claim" || v.ClaimID != "claim" {
		t.Fatalf("unexpected violation identity: %+v", v)
	}
	if len(v.RequiredBoundFiles) == 0 ||
		len(v.RequiredEvidence) == 0 ||
		len(v.RequiredReasoningEvidence) == 0 {
		t.Fatalf("violation missing required co-change scope: %+v", v)
	}
}

func TestEvidenceStatusFailsWhenImpactHasViolations(t *testing.T) {
	status := evidenceStatus(&impactEvidence{Violations: []impactViolation{{ClaimID: "claim"}}})
	if status != "failed" {
		t.Fatalf("status = %s, want failed", status)
	}
}
