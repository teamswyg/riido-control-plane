package main

type decisionSummary struct {
	RegisteredDecisionCount int            `json:"registered_decision_count"`
	ConsumedDecisionCount   int            `json:"consumed_decision_count"`
	DispositionCounts       []summaryCount `json:"disposition_counts"`
	PriorityCounts          []summaryCount `json:"priority_counts"`
	NextArtifactCounts      []summaryCount `json:"next_artifact_counts"`
	DecisionSourceCounts    []summaryCount `json:"decision_source_counts,omitempty"`
}

type summaryCount struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}
