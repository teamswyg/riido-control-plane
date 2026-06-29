package main

import (
	"slices"
	"testing"
)

func TestIgnoredCommandCandidateCarriesSubject(t *testing.T) {
	plan := dispatchPlan{
		GeneratedAt: "2026-06-25T00:00:00Z",
		ExpiresAt:   "2026-06-26T00:00:00Z",
		IgnoredCommands: []selectedRefreshCommand{{
			LoopID:      "control_plane_pressure_candidate",
			Kind:        "target_verifier",
			Command:     "go test ./tools/controlplaneperf",
			CandidateID: "candidate_one",
			SubjectKind: "claim_coverage_gap",
			ClaimIDs: []string{
				"control_plane_pressure_claim",
				"ai_thread_history_claim",
			},
			EvidenceChainIDs: []string{
				"target_verifier_chain",
				"control_plane_pressure_chain",
			},
		}},
	}
	item := candidateEvidenceFromPlan(plan).Candidates[0]
	if item.Subject == nil ||
		item.Subject.Kind != "loop_refresh_ignored_command" ||
		item.Subject.LoopID != "control_plane_pressure_candidate" ||
		item.Subject.CommandKind != "target_verifier" ||
		item.Subject.SourceCandidateID != "candidate_one" ||
		item.Subject.SourceSubjectKind != "claim_coverage_gap" {
		t.Fatalf("subject = %+v", item.Subject)
	}
	if !slices.Equal(item.Subject.ClaimIDs, []string{
		"ai_thread_history_claim",
		"control_plane_pressure_claim",
	}) {
		t.Fatalf("claim ids = %+v", item.Subject.ClaimIDs)
	}
	if !slices.Equal(item.Subject.EvidenceChainIDs, []string{
		"control_plane_pressure_chain",
		"target_verifier_chain",
	}) {
		t.Fatalf("evidence chain ids = %+v", item.Subject.EvidenceChainIDs)
	}
}

func TestStaleSourceCandidateCarriesSubject(t *testing.T) {
	plan := dispatchPlan{
		GeneratedAt: "2026-06-25T00:00:00Z",
		ExpiresAt:   "2026-06-26T00:00:00Z",
		Status:      "source_stale",
		SourceStaleSources: []staleRefreshSource{{
			SourcePath:  "stale.json",
			GeneratedAt: "2026-06-23T00:00:00Z",
			ExpiresAt:   "2026-06-24T00:00:00Z",
			Reason:      "expired",
		}},
	}
	item := candidateEvidenceFromPlan(plan).Candidates[0]
	if item.Subject == nil ||
		item.Subject.Kind != "loop_refresh_stale_source" ||
		item.Subject.SourcePath != "stale.json" ||
		item.Subject.Reason != "expired" {
		t.Fatalf("subject = %+v", item.Subject)
	}
}
