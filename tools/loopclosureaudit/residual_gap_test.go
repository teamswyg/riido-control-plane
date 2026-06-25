package main

import "testing"

func TestLoopClosureAuditEvidenceExposesResidualGaps(t *testing.T) {
	got := newEvidence(manifest{ResidualGaps: []residualGap{{
		ID:           "gap",
		Observation:  "observed",
		Risk:         "risk",
		NextLoop:     "closed_loop_candidate",
		NextArtifact: "candidate",
		NextCommand:  "go test ./tools/loopclosureaudit",
	}}})
	if got.ResidualGapCount != 1 || got.ResidualGaps[0].NextCommand == "" {
		t.Fatalf("expected residual gap evidence, got %+v", got)
	}
}

func TestLoopClosureAuditRejectsResidualGapWithoutCommand(t *testing.T) {
	m, deps := loadForTest(t)
	m.ResidualGaps = []residualGap{{
		ID:           "gap",
		Observation:  "observed",
		Risk:         "risk",
		NextLoop:     "closed_loop_candidate",
		NextArtifact: "candidate",
	}}
	if err := verifyAll("../..", m, deps); err == nil {
		t.Fatal("expected missing residual gap command to fail")
	}
}
