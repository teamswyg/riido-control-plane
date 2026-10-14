package main

import (
	"fmt"
	"strings"
)

const targetVerifierRunnableSummaryLimit = 2

func targetVerifierRunnableSummaryFor(
	plan *targetVerifierPlan,
	evidenceRef string,
) string {
	if plan == nil || len(plan.RunnableCommands) == 0 {
		return ""
	}
	if evidenceRef == "" {
		evidenceRef = "loop-registry-evidence"
	}
	limit := targetVerifierRunnableSummaryLimit
	if len(plan.RunnableCommands) < limit {
		limit = len(plan.RunnableCommands)
	}
	parts := append([]string(nil), plan.RunnableCommands[:limit]...)
	if remaining := len(plan.RunnableCommands) - limit; remaining > 0 {
		parts = append(parts,
			fmt.Sprintf("+%d more in %s", remaining, evidenceRef))
	}
	return "runnable: " + strings.Join(parts, " ; ")
}
