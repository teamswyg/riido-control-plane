package main

import "testing"

func TestBuildDispatchPlanRejectsCommandCountMismatch(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	_, err := buildDispatchPlan(repoRootForTest(t), refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		GeneratedAt:   "2026-06-24T00:00:00Z",
		ExpiresAt:     "2026-06-26T00:00:00Z",
		CommandCount:  2,
		Commands: []selectedRefreshCommand{{
			LoopID:  "closed_loop_candidate",
			Kind:    "refresh_workflow",
			Command: "gh workflow run loop-registry.yml --ref main",
		}},
	})
	if err == nil {
		t.Fatal("expected command count mismatch rejection")
	}
}

func TestBuildDispatchPlanRejectsRefreshRequiredWithoutCommands(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	_, err := buildDispatchPlan(repoRootForTest(t), refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		GeneratedAt:   "2026-06-24T00:00:00Z",
		ExpiresAt:     "2026-06-26T00:00:00Z",
	})
	if err == nil {
		t.Fatal("expected refresh_required command rejection")
	}
}
