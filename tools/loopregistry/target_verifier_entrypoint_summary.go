package main

import (
	"fmt"
	"strings"
)

const targetVerifierEntrypointSummaryLimit = 2

func targetVerifierEntrypointSummaryFor(
	plan *targetVerifierPlan,
	evidenceRef string,
) string {
	if plan == nil || len(plan.EntrypointCommands) == 0 {
		return ""
	}
	if evidenceRef == "" {
		evidenceRef = "loop-registry-evidence"
	}
	limit := targetVerifierEntrypointSummaryLimit
	if len(plan.EntrypointCommands) < limit {
		limit = len(plan.EntrypointCommands)
	}
	parts := append([]string(nil), plan.EntrypointCommands[:limit]...)
	if remaining := len(plan.EntrypointCommands) - limit; remaining > 0 {
		parts = append(parts,
			fmt.Sprintf("+%d more in %s", remaining, evidenceRef))
	}
	return "entrypoints: " + strings.Join(parts, " ; ")
}
