package main

import "testing"

func TestBuildDispatchPlanIncludesWeeklyOpenQuestionsRefreshWorkflow(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-07-04T00:00:00Z")
	got, err := buildDispatchPlan(repoRootForTest(t), refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		GeneratedAt:   "2026-07-03T00:00:00Z",
		ExpiresAt:     "2026-07-05T00:00:00Z",
		Commands: []selectedRefreshCommand{{
			LoopID:  "open_decision_queue",
			Kind:    "refresh_workflow",
			Command: "gh workflow run open-questions.yml --ref main",
		}, {
			LoopID:  "open_decision_queue",
			Kind:    "target_verifier",
			Command: "go test ./tools/openquestions",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "dispatch_required" || got.DispatchCount != 1 {
		t.Fatalf("plan status/count = %+v", got)
	}
	if got.Dispatches[0].WorkflowFile != "open-questions.yml" {
		t.Fatalf("dispatch = %+v", got.Dispatches[0])
	}
	if got.IgnoredCommandCount != 1 {
		t.Fatalf("ignored command count = %+v", got)
	}
}
