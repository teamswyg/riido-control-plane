package main

import "fmt"

func verifyDecisionNextArtifact(candidate closedLoopCandidate, decision decisionRecord) error {
	if !containsString(candidate.RequiredNextArtifacts, decision.NextArtifact) {
		return fmt.Errorf("candidate %s decision next_artifact %s is not required by candidate artifact",
			candidate.ID, decision.NextArtifact)
	}
	next, err := subjectNextArtifact(candidate)
	if err != nil {
		return err
	}
	if next != "" && decision.NextArtifact != next {
		return fmt.Errorf("candidate %s decision next_artifact %s does not match subject next_artifact %s",
			candidate.ID, decision.NextArtifact, next)
	}
	return nil
}
