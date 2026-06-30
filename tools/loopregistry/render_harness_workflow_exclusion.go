package main

import (
	"fmt"
	"strings"
)

func renderHarnessWorkflowExclusions(b *strings.Builder, exclusions []harnessWorkflowExclusion) {
	if len(exclusions) == 0 {
		return
	}
	fmt.Fprintln(b, "## Harness Workflow Exclusions")
	fmt.Fprintln(b)
	fmt.Fprintln(b, "| Workflow | Replacement claim | Replacement evidence | Reason |")
	fmt.Fprintln(b, "| --- | --- | --- | --- |")
	for _, exclusion := range exclusions {
		fmt.Fprintf(b, "| `%s` | `%s` | `%s` | %s |\n",
			exclusion.Workflow, exclusion.ReplacementClaim,
			exclusion.ReplacementEvidence, exclusion.Reason)
	}
	fmt.Fprintln(b)
}
