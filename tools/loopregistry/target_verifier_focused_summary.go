package main

import (
	"fmt"
	"strings"
)

const targetVerifierFocusedSummaryLimit = 2

func targetVerifierFocusedSummaryFor(
	plan *targetVerifierPlan,
	evidenceRef string,
) string {
	if plan == nil || len(plan.FocusedCommands) == 0 {
		return ""
	}
	if evidenceRef == "" {
		evidenceRef = "loop-registry-evidence"
	}
	limit := targetVerifierFocusedSummaryLimit
	if len(plan.FocusedCommands) < limit {
		limit = len(plan.FocusedCommands)
	}
	parts := append([]string(nil), plan.FocusedCommands[:limit]...)
	if remaining := len(plan.FocusedCommands) - limit; remaining > 0 {
		parts = append(parts,
			fmt.Sprintf("+%d more in %s", remaining, evidenceRef))
	}
	return "focused: " + strings.Join(parts, " ; ")
}
