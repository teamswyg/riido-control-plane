package main

import "testing"

func TestResidualGapCandidatesExportExecutablePlan(t *testing.T) {
	source := candidateSource{
		ID:                    "loop-closure-audit",
		SourceWorkflow:        ".github/workflows/loop-closure-audit.yml",
		SummaryArtifact:       "loop-closure-audit-evidence",
		CandidateArtifact:     "loop-closure-audit-closed-loop-candidates",
		HarnessLoop:           "loop_closure_audit",
		PromotionTarget:       "closed_loop_candidate",
		RequiredNextArtifacts: []string{"decision_record"},
	}
	gaps := []residualGap{{
		ID:           "gap",
		Observation:  "observed",
		Risk:         "risk",
		NextLoop:     "closed_loop_candidate_decision",
		NextArtifact: "decision_record",
		NextCommand:  "go run ./tools/closedloopcandidatedecision -check-doc",
	}}
	candidates := residualGapCandidates(source, gaps, "2026-06-24T12:00:00Z", "2026-06-25T12:00:00Z")
	if len(candidates) != 1 || candidates[0].AdoptionPlan[0].Command == "" {
		t.Fatalf("residual gap candidates = %+v", candidates)
	}
}
