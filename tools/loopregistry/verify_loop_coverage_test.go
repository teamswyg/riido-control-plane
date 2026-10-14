package main

import "testing"

func TestLoopVerifyCoverageRejectsUncoveredToken(t *testing.T) {
	m := manifest{
		Loops: []loopRecord{{ID: "loop-a", Verifies: []string{"thing_checked"}}},
		Claims: []claimBinding{{
			ID: "claim-a", Loop: "loop-a",
		}},
	}
	if err := verifyLoopVerifyCoverage(m); err == nil {
		t.Fatal("expected uncovered loop verify token failure")
	}
}

func TestLoopVerifyCoverageRejectsUnknownClaimToken(t *testing.T) {
	m := manifest{
		Loops: []loopRecord{{ID: "loop-a", Verifies: []string{"thing_checked"}}},
		Claims: []claimBinding{{
			ID: "claim-a", Loop: "loop-a", CoversVerifies: []string{"other_check"},
		}},
	}
	if err := verifyLoopVerifyCoverage(m); err == nil {
		t.Fatal("expected unknown claim verify token failure")
	}
}

func TestLoopVerifyCoverageAcceptsClaimCoveredToken(t *testing.T) {
	m := manifest{
		Loops: []loopRecord{{ID: "loop-a", Verifies: []string{"thing_checked"}}},
		Claims: []claimBinding{{
			ID: "claim-a", Loop: "loop-a", CoversVerifies: []string{"thing_checked"},
		}},
	}
	if err := verifyLoopVerifyCoverage(m); err != nil {
		t.Fatal(err)
	}
}
