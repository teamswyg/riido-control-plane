package main

import "testing"

func TestLoopEvidenceCoverageRejectsUncoveredSource(t *testing.T) {
	m := manifest{
		Loops: []loopRecord{{ID: "loop-a", Evidence: []evidenceSource{{Path: "artifact-a"}}}},
		Claims: []claimBinding{{
			ID: "claim-a", Loop: "loop-a",
		}},
	}
	if err := verifyLoopEvidenceCoverage(m); err == nil {
		t.Fatal("expected uncovered loop evidence source failure")
	}
}

func TestLoopEvidenceCoverageRejectsUnknownClaimSource(t *testing.T) {
	m := manifest{
		Loops: []loopRecord{{ID: "loop-a", Evidence: []evidenceSource{{Path: "artifact-a"}}}},
		Claims: []claimBinding{{
			ID: "claim-a", Loop: "loop-a", CoversEvidence: []string{"other-artifact"},
		}},
	}
	if err := verifyLoopEvidenceCoverage(m); err == nil {
		t.Fatal("expected unknown claim evidence source failure")
	}
}

func TestLoopEvidenceCoverageAcceptsClaimCoveredSource(t *testing.T) {
	m := manifest{
		Loops: []loopRecord{{ID: "loop-a", Evidence: []evidenceSource{{Path: "artifact-a"}}}},
		Claims: []claimBinding{{
			ID: "claim-a", Loop: "loop-a", CoversEvidence: []string{"artifact-a"},
		}},
	}
	if err := verifyLoopEvidenceCoverage(m); err != nil {
		t.Fatal(err)
	}
}
