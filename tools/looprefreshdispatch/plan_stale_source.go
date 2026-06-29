package main

func staleSourceDispatchPlan(source refreshCommandEvidence) dispatchPlan {
	generatedAt, expiresAt := evidenceWindow()
	return dispatchPlan{
		SchemaVersion:      dispatchPlanSchema,
		Status:             "source_stale",
		GeneratedAt:        generatedAt,
		ExpiresAt:          expiresAt,
		SourceStatus:       "source_stale",
		SourceStaleCount:   len(source.StaleSources),
		SourceStaleSources: source.StaleSources,
	}
}
