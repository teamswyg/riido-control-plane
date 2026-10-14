package main

func newDecisionSummary(m manifest, result verifyResult) decisionSummary {
	return decisionSummary{
		RegisteredDecisionCount: len(m.Decisions),
		ConsumedDecisionCount:   len(result.DecisionArtifacts),
		DispositionCounts:       decisionDispositionCounts(m.Decisions),
		PriorityCounts:          decisionPriorityCounts(m.Decisions),
		NextArtifactCounts:      decisionNextArtifactCounts(m.Decisions),
		DecisionSourceCounts:    decisionSourceCounts(result.DecisionArtifacts),
	}
}

func decisionDispositionCounts(decisions []decisionRecord) []summaryCount {
	return countDecisionValues(decisions, func(decision decisionRecord) string {
		return decision.Disposition
	})
}

func decisionPriorityCounts(decisions []decisionRecord) []summaryCount {
	return countDecisionValues(decisions, func(decision decisionRecord) string {
		return decision.Priority
	})
}

func decisionNextArtifactCounts(decisions []decisionRecord) []summaryCount {
	return countDecisionValues(decisions, func(decision decisionRecord) string {
		return decision.NextArtifact
	})
}
