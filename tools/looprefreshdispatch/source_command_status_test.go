package main

import "testing"

func TestBuildDispatchPlanAllowsFreshSourceWithoutDispatch(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	got, err := buildDispatchPlan(repoRootForTest(t), refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "fresh",
		GeneratedAt:   "2026-06-24T00:00:00Z",
		ExpiresAt:     "2026-06-26T00:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "no_dispatch_required" || got.DispatchCount != 0 {
		t.Fatalf("plan = %+v", got)
	}
}

func TestBuildDispatchPlanRejectsFreshSourceWithCommands(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	_, err := buildDispatchPlan(repoRootForTest(t), refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "fresh",
		GeneratedAt:   "2026-06-24T00:00:00Z",
		ExpiresAt:     "2026-06-26T00:00:00Z",
		CommandCount:  1,
		Commands: []selectedRefreshCommand{{
			LoopID:  "closed_loop_candidate",
			Kind:    "refresh_workflow",
			Command: "gh workflow run loop-registry.yml --ref main",
		}},
	})
	if err == nil {
		t.Fatal("expected fresh source command rejection")
	}
}
