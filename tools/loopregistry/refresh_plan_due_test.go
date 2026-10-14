package main

import "testing"

func TestLoopRegistryRefreshPlansExposeDueAndExpiry(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-24T00:00:00Z")
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll("../..", m, hashes)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil)
	for _, plan := range got.RefreshPlans {
		if plan.EvidenceGeneratedAt != "2026-06-24T00:00:00Z" {
			t.Fatalf("plan %s generated_at = %q", plan.LoopID, plan.EvidenceGeneratedAt)
		}
		if plan.NextRefreshDueAt == "" || plan.EvidenceExpiresAt == "" {
			t.Fatalf("plan %s missing due/expiry: %+v", plan.LoopID, plan)
		}
	}
}
