package main

import "testing"

func TestLoopRegistryEvidenceCarriesRefreshPlans(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll("../..", m, hashes)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil)
	if len(got.RefreshPlans) != len(m.Loops) {
		t.Fatalf("refresh plans = %d, loops = %d", len(got.RefreshPlans), len(m.Loops))
	}
	plan := got.RefreshPlans[0]
	if plan.LoopID == "" || plan.ManualRefreshCommand == "" {
		t.Fatalf("incomplete refresh plan: %+v", plan)
	}
	if len(plan.EvidenceArtifacts) == 0 {
		t.Fatalf("missing evidence artifacts: %+v", plan)
	}
}
