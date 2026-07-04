package main

import (
	"fmt"
	"strings"
)

func renderLoopCompletion(b *strings.Builder, items []loopCompletion) {
	summary := summarizeLoopCompletions(items)
	fmt.Fprintln(b, "## Loop Completion Gate")
	fmt.Fprintln(b)
	fmt.Fprintf(b, "- threshold basis points: `%d`\n", summary.ThresholdBasisPoints)
	fmt.Fprintf(b, "- minimum loop completion basis points: `%d`\n", summary.MinCompletionBasisPoints)
	fmt.Fprintf(b, "- verified loops: `%d` / `%d`\n", summary.VerifiedLoopCount, summary.LoopCount)
	fmt.Fprintf(b, "- below-threshold loops: `%d`\n\n", summary.BelowThresholdCount)
	fmt.Fprintln(b, "| Loop | Kind | Status | Passed | Required | Basis points | Missing |")
	fmt.Fprintln(b, "| --- | --- | --- | ---: | ---: | ---: | --- |")
	for _, item := range items {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` | `%d` | `%d` | `%d` | `%s` |\n",
			item.LoopID, item.Kind, item.Status, item.PassedChecks,
			item.RequiredChecks, item.CompletionBasisPoints,
			strings.Join(item.MissingChecks, ", "))
	}
	fmt.Fprintln(b)
}
