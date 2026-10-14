package main

import "testing"

func TestTargetVerifierPlanFocusesAddedClaims(t *testing.T) {
	impact := &impactEvidence{
		Enabled:      true,
		ChangedFiles: []string{"tools/a.go", "tools/b.go"},
		AddedClaims:  []impactClaim{{ID: "claim-b"}},
	}
	attachTargetVerifierPlan(impact, focusIndex(), focusSurfaces())
	plan := impact.TargetVerifierPlan
	if len(plan.FocusedClaimIDs) != 1 || plan.FocusedClaimIDs[0] != "claim-b" {
		t.Fatalf("focused claims = %#v", plan.FocusedClaimIDs)
	}
	if len(plan.FocusedCommands) != 1 || plan.FocusedCommands[0] != "test-b" {
		t.Fatalf("focused commands = %#v", plan.FocusedCommands)
	}
	if plan.CommandCount != 2 || plan.FocusedCommandCount != 1 {
		t.Fatalf("plan counts = %+v", plan)
	}
}

func TestTargetVerifierPlanFocusFallsBackToBoundSurfaces(t *testing.T) {
	impact := &impactEvidence{
		Enabled:       true,
		ChangedFiles:  []string{"tools/a.go", "tools/b.go"},
		BoundSurfaces: []impactBoundSurface{{ID: "claim-a"}},
	}
	attachTargetVerifierPlan(impact, focusIndex(), focusSurfaces())
	plan := impact.TargetVerifierPlan
	if len(plan.FocusedClaimIDs) != 1 || plan.FocusedClaimIDs[0] != "claim-a" {
		t.Fatalf("focused claims = %#v", plan.FocusedClaimIDs)
	}
	if len(plan.FocusedCommands) != 1 || plan.FocusedCommands[0] != "test-a" {
		t.Fatalf("focused commands = %#v", plan.FocusedCommands)
	}
}

func focusIndex() architectureIndex {
	return architectureIndex{Paths: []architecturePathBinding{
		focusPath("tools/a.go", "claim-a", "test-a"),
		focusPath("tools/b.go", "claim-b", "test-b"),
	}}
}

func focusPath(path, claimID, command string) architecturePathBinding {
	commands := appendUnique([]string{command}, "test-a", "test-b")
	return architecturePathBinding{
		Path:             path,
		Kind:             "code",
		LoopIDs:          []string{"loop-a"},
		ClaimIDs:         []string{claimID},
		VerifierCommands: commands,
		EvidenceChainIDs: []string{"chain-a"},
	}
}

func focusSurfaces() []claimSurface {
	return []claimSurface{
		{ID: "claim-a", VerifierCommands: []string{"test-a"}},
		{ID: "claim-b", VerifierCommands: []string{"test-b"}},
	}
}
