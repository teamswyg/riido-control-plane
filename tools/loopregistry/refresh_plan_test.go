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

func TestLoopRegistryEvidenceCarriesArtifactRefreshOwners(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll("../..", m, hashes)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil)
	for _, plan := range got.RefreshPlans {
		if len(plan.EvidenceArtifacts) != len(plan.EvidenceRefreshes) {
			t.Fatalf("plan %s artifact owners = %d, artifacts = %d",
				plan.LoopID, len(plan.EvidenceRefreshes), len(plan.EvidenceArtifacts))
		}
		for _, refresh := range plan.EvidenceRefreshes {
			if refresh.Artifact == "" || refresh.RefreshWorkflow == "" ||
				refresh.ManualRefreshCommand == "" {
				t.Fatalf("incomplete artifact refresh owner in plan %s: %+v", plan.LoopID, refresh)
			}
		}
	}
}
