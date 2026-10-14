package main

import (
	"fmt"
	"sort"
	"strings"
)

const targetVerifierAnnotationPathLimit = 2

func targetVerifierPathSummaryFor(
	plan *targetVerifierPlan,
	evidenceRef string,
) string {
	paths := targetVerifierPlanPaths(plan)
	if len(paths) == 0 {
		return ""
	}
	if evidenceRef == "" {
		evidenceRef = "loop-registry-evidence"
	}
	limit := targetVerifierAnnotationPathLimit
	if len(paths) < limit {
		limit = len(paths)
	}
	parts := append([]string(nil), paths[:limit]...)
	if remaining := len(paths) - limit; remaining > 0 {
		parts = append(parts, fmt.Sprintf("+%d more in %s", remaining, evidenceRef))
	}
	return "paths: " + strings.Join(parts, ", ")
}

func targetVerifierPlanPaths(plan *targetVerifierPlan) []string {
	if plan == nil {
		return nil
	}
	seen := map[string]struct{}{}
	for _, path := range plan.Paths {
		if path.Path != "" {
			seen[path.Path] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for path := range seen {
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}
