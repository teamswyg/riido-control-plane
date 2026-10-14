package main

import "testing"

func TestBuildDispatchPlanGroupsSafeWorkflowRuns(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	got, err := buildDispatchPlan(repoRootForTest(t), refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		GeneratedAt:   "2026-06-24T00:00:00Z",
		ExpiresAt:     "2026-06-26T00:00:00Z",
		CommandCount:  4,
		Commands: []selectedRefreshCommand{{
			LoopID:  "ai_thread_history",
			Kind:    "refresh_workflow",
			Command: "gh workflow run loop-registry.yml --ref main",
		}, {
			LoopID:  "closed_loop_candidate",
			Kind:    "refresh_workflow",
			Command: "gh workflow run loop-registry.yml --ref main",
		}, {
			LoopID:  "open_decision_queue",
			Kind:    "refresh_workflow",
			Command: "gh workflow run open-questions.yml --ref main",
		}, {
			LoopID:  "ai_thread_history",
			Kind:    "target_verifier",
			Command: "go test ./internal/riidoaiserver",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "dispatch_required" || got.DispatchCount != 2 {
		t.Fatalf("plan status/count = %+v", got)
	}
	if got.GeneratedAt != "2026-06-25T00:00:00Z" || got.ExpiresAt != "2026-06-26T00:00:00Z" {
		t.Fatalf("plan freshness = %+v", got)
	}
	if got.SourceGeneratedAt == "" || got.SourceExpiresAt == "" {
		t.Fatalf("source freshness missing = %+v", got)
	}
	if got.SourceCommandCount != 4 {
		t.Fatalf("source command count = %+v", got)
	}
	first := got.Dispatches[0]
	if first.WorkflowFile != "loop-registry.yml" || first.CommandCount != 2 {
		t.Fatalf("first dispatch = %+v", first)
	}
	if first.VerifiedCommand != "gh workflow run loop-registry.yml --ref main" {
		t.Fatalf("first verified command = %q", first.VerifiedCommand)
	}
	second := got.Dispatches[1]
	if second.WorkflowFile != "open-questions.yml" || second.CommandCount != 1 {
		t.Fatalf("second dispatch = %+v", second)
	}
	if second.VerifiedCommand != "gh workflow run open-questions.yml --ref main" {
		t.Fatalf("second verified command = %q", second.VerifiedCommand)
	}
	if got.IgnoredCommandCount != 1 || got.IgnoredCommandKinds[0] != "target_verifier" {
		t.Fatalf("ignored = %+v", got)
	}
	if len(got.IgnoredCommands) != 1 {
		t.Fatalf("ignored command details missing = %+v", got)
	}
	ignored := got.IgnoredCommands[0]
	if ignored.LoopID != "ai_thread_history" || ignored.Kind != "target_verifier" {
		t.Fatalf("ignored command identity = %+v", ignored)
	}
	if ignored.Command != "go test ./internal/riidoaiserver" {
		t.Fatalf("ignored command string = %q", ignored.Command)
	}
}
