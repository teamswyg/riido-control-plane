package main

import "testing"

func TestBuildDispatchPlanKeepsWorkflowInputs(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	got, err := buildDispatchPlan(repoRootForTest(t), refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		GeneratedAt:   "2026-06-24T00:00:00Z",
		ExpiresAt:     "2026-06-26T00:00:00Z",
		CommandCount:  1,
		Commands: []selectedRefreshCommand{{
			LoopID:  "operational_readiness_release_harness",
			Kind:    "refresh_workflow",
			Command: "gh workflow run ai-agent-client-testnet-load.yml -f scenario=public -f duration=120s -f concurrency=128",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.DispatchCount != 1 || got.IgnoredCommandCount != 0 {
		t.Fatalf("dispatch plan = %+v", got)
	}
	want := "gh workflow run ai-agent-client-testnet-load.yml --ref main -f scenario=public -f duration=120s -f concurrency=128"
	if got.Dispatches[0].VerifiedCommand != want {
		t.Fatalf("verified command = %q", got.Dispatches[0].VerifiedCommand)
	}
	inputs := got.Dispatches[0].Inputs
	if len(inputs) != 3 || inputs[2].Name != "concurrency" || inputs[2].Value != "128" {
		t.Fatalf("inputs = %+v", inputs)
	}
}
