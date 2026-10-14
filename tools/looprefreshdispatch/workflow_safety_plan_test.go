package main

import "testing"

func TestBuildDispatchPlanRejectsUnsafeRefreshWorkflow(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	_, err := buildDispatchPlan(repoRootForTest(t), refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		GeneratedAt:   "2026-06-24T00:00:00Z",
		ExpiresAt:     "2026-06-26T00:00:00Z",
		CommandCount:  1,
		Commands: []selectedRefreshCommand{{
			LoopID:  "loop",
			Kind:    "refresh_workflow",
			Command: "gh workflow run ../deploy.yml --ref main",
		}},
	})
	if err == nil {
		t.Fatal("expected unsafe refresh workflow rejection")
	}
}
