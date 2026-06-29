package main

import (
	"fmt"
	"sort"
	"strings"
)

const targetVerifierAnnotationComponentLimit = 2

func targetVerifierAnnotationSummary(plan *targetVerifierPlan) string {
	parts := []string{}
	if componentSummary := targetVerifierComponentSummaryFor(
		plan, "loop-registry-evidence",
	); componentSummary != "" {
		parts = append(parts, componentSummary)
	}
	if loopSummary := targetVerifierLoopSummaryFor(
		plan, "loop-registry-evidence",
	); loopSummary != "" {
		parts = append(parts, loopSummary)
	}
	if commandSummary := targetVerifierCommandSummary(plan); commandSummary != "" {
		parts = append(parts, commandSummary)
	}
	return strings.Join(parts, " ; ")
}

func targetVerifierComponentSummaryFor(
	plan *targetVerifierPlan,
	evidenceRef string,
) string {
	names := targetVerifierPlanComponentNames(plan)
	if len(names) == 0 {
		return ""
	}
	if evidenceRef == "" {
		evidenceRef = "loop-registry-evidence"
	}
	limit := targetVerifierAnnotationComponentLimit
	if len(names) < limit {
		limit = len(names)
	}
	parts := append([]string(nil), names[:limit]...)
	if remaining := len(names) - limit; remaining > 0 {
		parts = append(parts, fmt.Sprintf("+%d more in %s", remaining, evidenceRef))
	}
	return "components: " + strings.Join(parts, ", ")
}

func targetVerifierPlanComponentNames(plan *targetVerifierPlan) []string {
	if plan == nil {
		return nil
	}
	names := make([]string, 0, len(plan.Components))
	for _, component := range plan.Components {
		names = append(names, component.Component)
	}
	sort.Strings(names)
	return names
}
