package main

import (
	"fmt"
	"strings"
)

func renderPartialPromotion(b *strings.Builder, promotion partialPromotion) {
	b.WriteString("## Partial Promotion\n\n")
	fmt.Fprintf(b, "- candidate artifact: `%s`\n", promotion.CandidateArtifact)
	fmt.Fprintf(b, "- candidate count: `%d`\n", promotion.CandidateCount)
	fmt.Fprintf(b, "- stale partial count: `%d`\n", promotion.StalePartialCount)
	fmt.Fprintf(b, "- stale after days: `%d`\n\n", promotion.StaleAfterDays)
	if len(promotion.CandidateIDs) == 0 {
		b.WriteString("No stale partial candidates.\n\n")
		return
	}
	b.WriteString("| Candidate | Stale Partial |\n")
	b.WriteString("| --- | --- |\n")
	for i, candidateID := range promotion.CandidateIDs {
		fmt.Fprintf(b, "| `%s` | `%s` |\n", candidateID, promotion.StalePartialIDs[i])
	}
	b.WriteString("\n")
}
