package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func writeTargetVerifierSummary(
	w io.Writer,
	impact *impactEvidence,
	evidenceOut string,
) {
	message := targetVerifierSummary(impact, evidenceOut)
	if message == "" {
		return
	}
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintln(w, message)
}

func targetVerifierSummary(
	impact *impactEvidence,
	evidenceOut string,
) string {
	if impact == nil || impact.TargetVerifierPlan == nil {
		return ""
	}
	plan := impact.TargetVerifierPlan
	parts := []string{fmt.Sprintf(
		"riido target verifier plan: %d changed paths, %d matched paths, %d components, %d commands",
		plan.ChangedPathCount,
		plan.MatchedPathCount,
		plan.ComponentCount,
		plan.CommandCount,
	)}
	if componentSummary := targetVerifierComponentSummaryFor(
		plan, evidenceOut,
	); componentSummary != "" {
		parts = append(parts, "riido target verifier "+componentSummary)
	}
	if loopSummary := targetVerifierLoopSummaryFor(plan, evidenceOut); loopSummary != "" {
		parts = append(parts, "riido target verifier "+loopSummary)
	}
	if commandSummary := targetVerifierCommandSummaryFor(
		plan, evidenceOut,
	); commandSummary != "" {
		parts = append(parts, "riido target verifier "+commandSummary)
	}
	if evidenceOut != "" {
		parts = append(parts, "riido target verifier full_plan: "+evidenceOut)
	}
	return strings.Join(parts, "\n")
}
