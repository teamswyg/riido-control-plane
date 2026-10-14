package main

import "testing"

func TestLoopFailureCoverageRejectsUncoveredToken(t *testing.T) {
	m := manifest{
		Loops: []loopRecord{{ID: "loop-a", FailsWhen: []string{"thing_failed"}}},
		Claims: []claimBinding{{
			ID: "claim-a", Loop: "loop-a",
		}},
	}
	if err := verifyLoopFailureCoverage(m); err == nil {
		t.Fatal("expected uncovered loop failure token failure")
	}
}

func TestLoopFailureCoverageRejectsUnknownClaimToken(t *testing.T) {
	m := manifest{
		Loops: []loopRecord{{ID: "loop-a", FailsWhen: []string{"thing_failed"}}},
		Claims: []claimBinding{{
			ID: "claim-a", Loop: "loop-a", CoversFails: []string{"other_failure"},
		}},
	}
	if err := verifyLoopFailureCoverage(m); err == nil {
		t.Fatal("expected unknown claim failure token failure")
	}
}

func TestLoopFailureCoverageAcceptsClaimCoveredToken(t *testing.T) {
	m := manifest{
		Loops: []loopRecord{{ID: "loop-a", FailsWhen: []string{"thing_failed"}}},
		Claims: []claimBinding{{
			ID: "claim-a", Loop: "loop-a", CoversFails: []string{"thing_failed"},
		}},
	}
	if err := verifyLoopFailureCoverage(m); err != nil {
		t.Fatal(err)
	}
}
