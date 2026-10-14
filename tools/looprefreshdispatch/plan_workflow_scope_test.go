package main

import (
	"slices"
	"testing"
)

func TestWorkflowDispatchPreservesSemanticScope(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	got, err := buildDispatchPlan(repoRootForTest(t), workflowScopeInput())
	if err != nil {
		t.Fatal(err)
	}
	dispatch := got.Dispatches[0]
	if dispatch.WorkflowFile != "loop-registry.yml" || dispatch.CommandCount != 2 {
		t.Fatalf("dispatch identity = %+v", dispatch)
	}
	if !slices.Equal(dispatch.LoopIDs, []string{"ai_thread_history", "closed_loop_candidate"}) {
		t.Fatalf("loop ids = %+v", dispatch.LoopIDs)
	}
	if !slices.Equal(dispatch.ClaimIDs, []string{"claim_one", "claim_two"}) {
		t.Fatalf("claim ids = %+v", dispatch.ClaimIDs)
	}
	if !slices.Equal(dispatch.EvidenceChainIDs, []string{"chain_one", "chain_two"}) {
		t.Fatalf("evidence chain ids = %+v", dispatch.EvidenceChainIDs)
	}
}

func workflowScopeInput() refreshCommandEvidence {
	return refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		GeneratedAt:   "2026-06-24T00:00:00Z",
		ExpiresAt:     "2026-06-26T00:00:00Z",
		CommandCount:  2,
		Commands: []selectedRefreshCommand{{
			LoopID:           "closed_loop_candidate",
			Kind:             "refresh_workflow",
			Command:          "gh workflow run loop-registry.yml --ref main",
			ClaimIDs:         []string{"claim_two"},
			EvidenceChainIDs: []string{"chain_two"},
		}, {
			LoopID:           "ai_thread_history",
			Kind:             "refresh_workflow",
			Command:          "gh workflow run loop-registry.yml --ref main",
			ClaimIDs:         []string{"claim_one"},
			EvidenceChainIDs: []string{"chain_one"},
		}},
	}
}
