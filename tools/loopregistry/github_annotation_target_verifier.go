package main

import (
	"fmt"
	"io"
	"strings"
)

const targetVerifierAnnotationCommandLimit = 2

func writeTargetVerifierAnnotation(
	w io.Writer,
	impact *impactEvidence,
) {
	if impact == nil || impact.TargetVerifierPlan == nil {
		return
	}
	message := targetVerifierAnnotationSummary(impact.TargetVerifierPlan)
	if message == "" {
		return
	}
	fmt.Fprintf(
		w,
		"::notice title=%s::%s\n",
		githubAnnotationProperty("Riido target verifier plan"),
		githubAnnotationMessage(message),
	)
}

func targetVerifierCommandSummary(plan *targetVerifierPlan) string {
	return targetVerifierCommandSummaryFor(plan, "loop-registry-evidence")
}

func targetVerifierCommandSummaryFor(
	plan *targetVerifierPlan,
	evidenceRef string,
) string {
	if plan == nil || len(plan.VerifierCommands) == 0 {
		return ""
	}
	if evidenceRef == "" {
		evidenceRef = "loop-registry-evidence"
	}
	limit := targetVerifierAnnotationCommandLimit
	if len(plan.VerifierCommands) < limit {
		limit = len(plan.VerifierCommands)
	}
	parts := append([]string(nil), plan.VerifierCommands[:limit]...)
	remaining := plan.CommandCount - limit
	if remaining > 0 {
		parts = append(parts,
			fmt.Sprintf("+%d more in %s", remaining, evidenceRef))
	}
	return "commands: " + strings.Join(parts, " ; ")
}
