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

func TestLoopRegistryRefreshPlanCommandsExposeScope(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll("../..", m, hashes)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil)
	for _, plan := range got.RefreshPlans {
		if len(plan.NextCommands[0].ClaimIDs) == 0 {
			t.Fatalf("plan %s refresh command missing claim scope", plan.LoopID)
		}
		for _, command := range plan.NextCommands[1:] {
			if len(command.ClaimIDs) == 0 || len(command.EvidenceChainIDs) == 0 {
				t.Fatalf("plan %s scoped command = %+v", plan.LoopID, command)
			}
		}
	}
}
