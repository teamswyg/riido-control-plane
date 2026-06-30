package main

func decisionSourceCounts(artifacts []decisionArtifactEvidence) []summaryCount {
	counts := map[string]int{}
	for _, artifact := range artifacts {
		if artifact.DecisionSource == "" {
			continue
		}
		counts[artifact.DecisionSource]++
	}
	return sortedSummaryCounts(counts)
}
