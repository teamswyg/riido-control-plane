package main

import (
	"fmt"
	"strings"
)

func renderCompletion(b *strings.Builder, c completionEvidence) {
	b.WriteString("## Internal Completion Gate\n\n")
	fmt.Fprintf(b, "- status: `%s`\n", c.Status)
	fmt.Fprintf(b, "- threshold basis points: `%d`\n", c.ThresholdBasisPoints)
	fmt.Fprintf(b, "- internal completeness basis points: `%d`\n", c.InternalCompletenessBasisPoints)
	fmt.Fprintf(b, "- internal checks: `%d`\n", c.InternalCheckCount)
	fmt.Fprintf(b, "- internal covered: `%d`\n", c.InternalCoveredCount)
	fmt.Fprintf(b, "- internal partial: `%d`\n", c.InternalPartialCount)
	fmt.Fprintf(b, "- external excluded: `%d`\n", c.ExternalExcludedCount)
	fmt.Fprintf(b, "- external partial: `%d`\n\n", c.ExternalPartialCount)
	if len(c.ExternalExcludedChecks) == 0 {
		return
	}
	b.WriteString("External excluded checks:\n")
	for _, id := range c.ExternalExcludedChecks {
		fmt.Fprintf(b, "- `%s`\n", id)
	}
	b.WriteString("\n")
}
