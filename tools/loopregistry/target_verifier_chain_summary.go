package main

import (
	"fmt"
	"sort"
	"strings"
)

const targetVerifierAnnotationChainLimit = 2

func targetVerifierChainSummaryFor(
	plan *targetVerifierPlan,
	evidenceRef string,
) string {
	chains := targetVerifierPlanChainIDs(plan)
	if len(chains) == 0 {
		return ""
	}
	if evidenceRef == "" {
		evidenceRef = "loop-registry-evidence"
	}
	limit := targetVerifierAnnotationChainLimit
	if len(chains) < limit {
		limit = len(chains)
	}
	parts := append([]string(nil), chains[:limit]...)
	if remaining := len(chains) - limit; remaining > 0 {
		parts = append(parts, fmt.Sprintf("+%d more in %s", remaining, evidenceRef))
	}
	return "chains: " + strings.Join(parts, ", ")
}

func targetVerifierPlanChainIDs(plan *targetVerifierPlan) []string {
	if plan == nil {
		return nil
	}
	seen := map[string]struct{}{}
	for _, component := range plan.Components {
		for _, chainID := range component.EvidenceChainIDs {
			seen[chainID] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for chainID := range seen {
		out = append(out, chainID)
	}
	sort.Strings(out)
	return out
}
