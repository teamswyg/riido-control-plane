package main

import (
	"fmt"
	"strings"
)

func renderClaimCoverageGaps(b *strings.Builder, gaps []claimCoverageGap) {
	b.WriteString("## Claim Coverage Gaps\n\n")
	if len(gaps) == 0 {
		b.WriteString("No claim coverage token gaps detected.\n\n")
		return
	}
	b.WriteString("| Claim | Loop | Missing Dimensions |\n| --- | --- | --- |\n")
	for _, gap := range gaps {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` |\n",
			gap.ClaimID, gap.Loop, strings.Join(gap.MissingDimensions, ", "))
	}
	b.WriteString("\n")
}
