package main

import (
	"slices"
	"testing"
)

func TestIgnoredCommandPreservesSemanticScope(t *testing.T) {
	t.Setenv("RIIDO_EVIDENCE_NOW", "2026-06-25T00:00:00Z")
	got, err := buildDispatchPlan(repoRootForTest(t), refreshCommandEvidence{
		SchemaVersion: refreshCommandsSchema,
		Status:        "refresh_required",
		GeneratedAt:   "2026-06-24T00:00:00Z",
		ExpiresAt:     "2026-06-26T00:00:00Z",
		CommandCount:  1,
		Commands: []selectedRefreshCommand{{
			LoopID:                      "ai_thread_history",
			Kind:                        "target_verifier",
			Command:                     "go test ./internal/riidoaiserver",
			DecisionSource:              "template",
			DecisionTemplateSubjectKind: "loop_refresh_ignored_command",
			ClaimIDs:                    []string{"claim_two", "claim_one"},
			EvidenceChainIDs:            []string{"chain_one"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	ignored := got.IgnoredCommands[0]
	if !slices.Equal(ignored.ClaimIDs, []string{"claim_two", "claim_one"}) {
		t.Fatalf("claim ids = %+v", ignored.ClaimIDs)
	}
	if !slices.Equal(ignored.EvidenceChainIDs, []string{"chain_one"}) {
		t.Fatalf("evidence chain ids = %+v", ignored.EvidenceChainIDs)
	}
	if ignored.DecisionSource != "template" ||
		ignored.DecisionTemplateSubjectKind != "loop_refresh_ignored_command" {
		t.Fatalf("decision provenance = %+v", ignored)
	}
}
