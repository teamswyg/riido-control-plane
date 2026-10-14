package main

import "testing"

func TestLoopObservationCoverageRejectsUncoveredToken(t *testing.T) {
	m := manifest{
		Loops: []loopRecord{{ID: "loop-a", Observes: []string{"thing_seen"}}},
		Claims: []claimBinding{{
			ID: "claim-a", Loop: "loop-a",
		}},
	}
	if err := verifyLoopObservationCoverage(m); err == nil {
		t.Fatal("expected uncovered loop observe token failure")
	}
}

func TestLoopObservationCoverageRejectsUnknownClaimToken(t *testing.T) {
	m := manifest{
		Loops: []loopRecord{{ID: "loop-a", Observes: []string{"thing_seen"}}},
		Claims: []claimBinding{{
			ID: "claim-a", Loop: "loop-a", CoversObserves: []string{"other_observation"},
		}},
	}
	if err := verifyLoopObservationCoverage(m); err == nil {
		t.Fatal("expected unknown claim observe token failure")
	}
}

func TestLoopObservationCoverageAcceptsClaimCoveredToken(t *testing.T) {
	m := manifest{
		Loops: []loopRecord{{ID: "loop-a", Observes: []string{"thing_seen"}}},
		Claims: []claimBinding{{
			ID: "claim-a", Loop: "loop-a", CoversObserves: []string{"thing_seen"},
		}},
	}
	if err := verifyLoopObservationCoverage(m); err != nil {
		t.Fatal(err)
	}
}
