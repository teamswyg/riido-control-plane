package main

import (
	"fmt"
	"strings"
)

func verifyCandidateItem(m manifest, item closedLoopCandidate) (intakeSource, error) {
	if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Observation) == "" {
		return intakeSource{}, fmt.Errorf("candidate item must bind id and observation")
	}
	source, ok := findSourceForCandidate(m.Sources, item)
	if !ok {
		return intakeSource{}, fmt.Errorf("candidate %s targets unknown harness/loop edge", item.ID)
	}
	if err := verifyCandidatePromotionEdge(item, source); err != nil {
		return intakeSource{}, err
	}
	if err := verifyRequiredNextArtifacts(item.RequiredNextArtifacts, source.ID); err != nil {
		return intakeSource{}, err
	}
	if err := verifyAdoptionPlan(item); err != nil {
		return intakeSource{}, err
	}
	return source, nil
}

func findSourceForCandidate(sources []intakeSource, item closedLoopCandidate) (intakeSource, bool) {
	for _, source := range sources {
		if source.PromotionTarget == item.PromotionTarget && source.HarnessLoop == item.HarnessLoop {
			return source, true
		}
	}
	return intakeSource{}, false
}
