package main

import "sort"

func summarizeLoopCompletions(items []loopCompletion) loopCompletionSummary {
	summary := loopCompletionSummary{
		LoopCount:            len(items),
		ThresholdBasisPoints: loopCompletionThresholdBasisPoints,
	}
	if len(items) == 0 {
		return summary
	}
	summary.MinCompletionBasisPoints = 10000
	for _, item := range items {
		if item.CompletionBasisPoints < summary.MinCompletionBasisPoints {
			summary.MinCompletionBasisPoints = item.CompletionBasisPoints
		}
		if item.CompletionBasisPoints >= loopCompletionThresholdBasisPoints {
			summary.VerifiedLoopCount++
			continue
		}
		summary.BelowThresholdCount++
		summary.BelowThresholdLoopIDs = append(
			summary.BelowThresholdLoopIDs, item.LoopID)
	}
	sort.Strings(summary.BelowThresholdLoopIDs)
	return summary
}

func missingCompletionChecks(checks map[string]bool) []string {
	var out []string
	for name, ok := range checks {
		if !ok {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

func loopGraphCoverage(edges []graphEdge) map[string]bool {
	out := map[string]bool{}
	for _, edge := range edges {
		out[edge.From] = true
		out[edge.To] = true
	}
	return out
}
