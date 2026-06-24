package main

import (
	"fmt"
	"strings"
)

func verifyCandidateItem(m manifest, item closedLoopCandidate) error {
	if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Observation) == "" {
		return fmt.Errorf("candidate item must bind id and observation")
	}
	source, ok := findSourceForTarget(m.Sources, item.PromotionTarget)
	if !ok {
		return fmt.Errorf("candidate %s targets unknown loop %s", item.ID, item.PromotionTarget)
	}
	if item.HarnessLoop == "" {
		return fmt.Errorf("candidate %s must name harness loop", item.ID)
	}
	if err := verifyRequiredNextArtifacts(item.RequiredNextArtifacts, source.ID); err != nil {
		return err
	}
	return verifyAdoptionPlan(item)
}

func findSourceForTarget(sources []intakeSource, target string) (intakeSource, bool) {
	for _, source := range sources {
		if source.PromotionTarget == target {
			return source, true
		}
	}
	return intakeSource{}, false
}
