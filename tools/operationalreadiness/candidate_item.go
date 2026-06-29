package main

func newStalePartialCandidate(
	source producerSource,
	partial partialCheck,
	generatedAt string,
	expiresAt string,
	run runRecord,
) closedLoopCandidate {
	return closedLoopCandidate{
		ID:                    source.ID + ":" + partial.ID,
		SourceRef:             readinessSourceRef(source, generatedAt, expiresAt, run),
		Subject:               partialSubject(partial),
		HarnessLoop:           source.HarnessLoop,
		PromotionTarget:       source.PromotionTarget,
		PromotionEdge:         readinessPromotionEdge(source),
		Observation:           partialObservation(partial),
		Hypothesis:            partialHypothesis(partial),
		RequiredNextArtifacts: partialRequiredArtifacts(source, partial),
		AdoptionPlan:          partialAdoptionPlan(source, partial),
	}
}

func partialObservation(partial partialCheck) string {
	return "Operational readiness check " + partial.ID + " is stale partial evidence."
}

func partialHypothesis(partial partialCheck) string {
	return "Closing " + partial.NextArtifact + " can turn " + partial.Category +
		" readiness from partial into covered evidence."
}

func readinessPromotionEdge(source producerSource) graphEdge {
	return graphEdge{
		From:     source.HarnessLoop,
		To:       source.PromotionTarget,
		Relation: "promotes_failure_to",
	}
}
