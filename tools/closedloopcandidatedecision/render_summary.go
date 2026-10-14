package main

import (
	"fmt"
	"strings"
)

func renderDecisionSummary(b *strings.Builder, summary decisionSummary) {
	fmt.Fprintln(b, "## Decision Summary")
	fmt.Fprintln(b)
	fmt.Fprintf(b, "- registered decisions: `%d`\n", summary.RegisteredDecisionCount)
	fmt.Fprintf(b, "- consumed decisions: `%d`\n", summary.ConsumedDecisionCount)
	renderSummaryCounts(b, "disposition", summary.DispositionCounts)
	renderSummaryCounts(b, "priority", summary.PriorityCounts)
	renderSummaryCounts(b, "next artifact", summary.NextArtifactCounts)
	renderSummaryCounts(b, "decision source", summary.DecisionSourceCounts)
	fmt.Fprintln(b)
}

func renderSummaryCounts(b *strings.Builder, label string, counts []summaryCount) {
	if len(counts) == 0 {
		return
	}
	fmt.Fprintf(b, "- %s counts: %s\n", label, summaryCountText(counts))
}
