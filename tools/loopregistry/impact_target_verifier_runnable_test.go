package main

import "testing"

func TestTargetVerifierPlanExposesFocusedRunnableCommands(t *testing.T) {
	impact := &impactEvidence{
		Enabled:      true,
		ChangedFiles: []string{"tools/a.go", "tools/b.go"},
		AddedClaims:  []impactClaim{{ID: "claim-b"}},
	}
	attachTargetVerifierPlan(impact, focusIndex(), focusSurfaces())
	plan := impact.TargetVerifierPlan
	if plan.RunnableCommandCount != 1 ||
		len(plan.RunnableCommands) != 1 ||
		plan.RunnableCommands[0] != "test-b" {
		t.Fatalf("runnable commands = %+v", plan)
	}
}

func TestTargetVerifierPlanFallsBackToEntrypointRunnableCommands(t *testing.T) {
	impact := &impactEvidence{
		Enabled:      true,
		ChangedFiles: []string{"tools/a.go", "tools/b.go"},
	}
	attachTargetVerifierPlan(impact, focusIndex(), focusSurfaces())
	plan := impact.TargetVerifierPlan
	if plan.RunnableCommandCount != len(plan.EntrypointCommands) ||
		len(plan.RunnableCommands) != len(plan.EntrypointCommands) {
		t.Fatalf("runnable commands = %+v", plan)
	}
}
