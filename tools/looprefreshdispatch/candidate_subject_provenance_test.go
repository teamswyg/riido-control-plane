package main

import "testing"

func TestIgnoredCommandCandidateCarriesDecisionProvenance(t *testing.T) {
	plan := dispatchPlan{
		GeneratedAt: "2026-06-25T00:00:00Z",
		ExpiresAt:   "2026-06-26T00:00:00Z",
		IgnoredCommands: []selectedRefreshCommand{{
			LoopID:                      "control_plane_pressure_candidate",
			Kind:                        "target_verifier",
			Command:                     "go test ./tools/controlplaneperf",
			CandidateID:                 "candidate_one",
			SubjectKind:                 "loop_refresh_ignored_command",
			DecisionSource:              "template",
			DecisionTemplateSubjectKind: "loop_refresh_ignored_command",
		}},
	}
	subject := candidateEvidenceFromPlan(plan).Candidates[0].Subject
	if subject.DecisionSource != "template" ||
		subject.DecisionTemplateSubjectKind != "loop_refresh_ignored_command" {
		t.Fatalf("decision provenance = %+v", subject)
	}
}
