package main

import "strings"

func newCandidate(source promotionSource, claimID, summary string) closedLoopCandidate {
	return closedLoopCandidate{
		ID:                    source.ID + ":" + sanitizeID(claimID),
		HarnessLoop:           source.HarnessLoop,
		PromotionTarget:       source.PromotionTarget,
		Observation:           "Harness " + source.ID + " reported unverified claim " + claimID + ".",
		Hypothesis:            summary,
		RequiredNextArtifacts: append([]string(nil), source.RequiredNextArtifacts...),
		AdoptionPlan:          adoptionPlan(source),
	}
}

func sanitizeID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "workflow"
	}
	return strings.NewReplacer(" ", "_", "/", "_", ":", "_").Replace(value)
}
