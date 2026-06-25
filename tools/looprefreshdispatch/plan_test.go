package main

import "testing"

func TestBuildDispatchPlanGroupsSafeWorkflowRuns(t *testing.T) {
	got, err := buildDispatchPlan(repoRootForTest(t), refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		Commands: []selectedRefreshCommand{{
			LoopID:  "ai_thread_history",
			Kind:    "refresh_workflow",
			Command: "gh workflow run loop-registry.yml --ref main",
		}, {
			LoopID:  "closed_loop_candidate",
			Kind:    "refresh_workflow",
			Command: "gh workflow run loop-registry.yml --ref main",
		}, {
			LoopID:  "ai_thread_history",
			Kind:    "target_verifier",
			Command: "go test ./internal/riidoaiserver",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "dispatch_required" || got.DispatchCount != 1 {
		t.Fatalf("plan status/count = %+v", got)
	}
	dispatch := got.Dispatches[0]
	if dispatch.WorkflowFile != "loop-registry.yml" || dispatch.CommandCount != 2 {
		t.Fatalf("dispatch = %+v", dispatch)
	}
	if got.IgnoredCommandCount != 1 || got.IgnoredCommandKinds[0] != "target_verifier" {
		t.Fatalf("ignored = %+v", got)
	}
}

func TestBuildDispatchPlanRejectsUnsafeRefreshWorkflow(t *testing.T) {
	_, err := buildDispatchPlan(repoRootForTest(t), refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
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
