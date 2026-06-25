package main

import "testing"

func TestLoopRegistryRefreshPlansExposeOrderedNextCommands(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll("../..", m, hashes)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil)
	for _, plan := range got.RefreshPlans {
		if len(plan.NextCommands) != len(plan.VerifierCommands)+1 {
			t.Fatalf("plan %s next commands = %+v", plan.LoopID, plan.NextCommands)
		}
		if plan.NextCommands[0].Kind != "refresh_workflow" {
			t.Fatalf("plan %s first command = %+v", plan.LoopID, plan.NextCommands[0])
		}
	}
}
