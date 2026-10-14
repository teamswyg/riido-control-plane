package main

import (
	"fmt"
	"sort"
	"strings"
)

const targetVerifierAnnotationClaimLimit = 2

func targetVerifierClaimSummaryFor(
	plan *targetVerifierPlan,
	evidenceRef string,
) string {
	claims := targetVerifierPlanClaimIDs(plan)
	if len(claims) == 0 {
		return ""
	}
	if evidenceRef == "" {
		evidenceRef = "loop-registry-evidence"
	}
	limit := targetVerifierAnnotationClaimLimit
	if len(claims) < limit {
		limit = len(claims)
	}
	parts := append([]string(nil), claims[:limit]...)
	if remaining := len(claims) - limit; remaining > 0 {
		parts = append(parts, fmt.Sprintf("+%d more in %s", remaining, evidenceRef))
	}
	return "claims: " + strings.Join(parts, ", ")
}

func targetVerifierPlanClaimIDs(plan *targetVerifierPlan) []string {
	if plan == nil {
		return nil
	}
	seen := map[string]struct{}{}
	for _, component := range plan.Components {
		for _, claimID := range component.ClaimIDs {
			seen[claimID] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for claimID := range seen {
		out = append(out, claimID)
	}
	sort.Strings(out)
	return out
}
