package main

import (
	"fmt"
	"sort"
	"strings"
)

const targetVerifierAnnotationLoopLimit = 2

func targetVerifierLoopSummaryFor(
	plan *targetVerifierPlan,
	evidenceRef string,
) string {
	loops := targetVerifierPlanLoopIDs(plan)
	if len(loops) == 0 {
		return ""
	}
	if evidenceRef == "" {
		evidenceRef = "loop-registry-evidence"
	}
	limit := targetVerifierAnnotationLoopLimit
	if len(loops) < limit {
		limit = len(loops)
	}
	parts := append([]string(nil), loops[:limit]...)
	if remaining := len(loops) - limit; remaining > 0 {
		parts = append(parts, fmt.Sprintf("+%d more in %s", remaining, evidenceRef))
	}
	return "loops: " + strings.Join(parts, ", ")
}

func targetVerifierPlanLoopIDs(plan *targetVerifierPlan) []string {
	if plan == nil {
		return nil
	}
	seen := map[string]struct{}{}
	for _, component := range plan.Components {
		for _, loopID := range component.LoopIDs {
			seen[loopID] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for loopID := range seen {
		out = append(out, loopID)
	}
	sort.Strings(out)
	return out
}
