package main

import "testing"

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
