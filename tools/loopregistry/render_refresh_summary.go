package main

import (
	"fmt"
	"strings"
)

func renderRefreshPlanSummary(b *strings.Builder, summary refreshPlanSummary) {
	fmt.Fprintln(b, "## Refresh Plan Summary")
	fmt.Fprintln(b)
	fmt.Fprintf(b, "- plans: `%d`\n", summary.PlanCount)
	fmt.Fprintf(b, "- refresh workflows: `%d`\n", summary.RefreshWorkflowCount)
	fmt.Fprintf(b, "- evidence artifacts: `%d`\n", summary.EvidenceArtifactCount)
	fmt.Fprintf(b, "- next commands: `%d`\n", summary.NextCommandCount)
	fmt.Fprintf(b, "- verifier commands: `%d`\n", summary.VerifierCommandCount)
	fmt.Fprintf(b, "- claim bindings: `%d`\n", summary.ClaimBindingCount)
	fmt.Fprintf(b, "- manual commands: `%d`\n\n", summary.ManualCommandCount)
}
