package main

import "fmt"

func verifyDecisionNextArtifact(candidate closedLoopCandidate, decision decisionRecord) error {
	if !containsString(candidate.RequiredNextArtifacts, decision.NextArtifact) {
		return fmt.Errorf("candidate %s decision next_artifact %s is not required by candidate artifact",
			candidate.ID, decision.NextArtifact)
	}
	return nil
}
