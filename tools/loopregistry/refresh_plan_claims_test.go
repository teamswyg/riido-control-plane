package main

import "testing"

func TestLoopRegistryRefreshPlansExposeClaimVerifiers(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll("../..", m, hashes)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil)
	for _, plan := range got.RefreshPlans {
		if len(plan.ClaimIDs) == 0 {
			t.Fatalf("plan %s missing claim ids: %+v", plan.LoopID, plan)
		}
		if len(plan.VerifierCommands) == 0 {
			t.Fatalf("plan %s missing verifier commands: %+v", plan.LoopID, plan)
		}
	}
}
