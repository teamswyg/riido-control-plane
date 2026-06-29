package main

import "strings"

func targetVerifierAnnotationSummary(plan *targetVerifierPlan) string {
	parts := []string{}
	if pathSummary := targetVerifierPathSummaryFor(
		plan, "loop-registry-evidence",
	); pathSummary != "" {
		parts = append(parts, pathSummary)
	}
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
	if claimSummary := targetVerifierClaimSummaryFor(
		plan, "loop-registry-evidence",
	); claimSummary != "" {
		parts = append(parts, claimSummary)
	}
	if chainSummary := targetVerifierChainSummaryFor(
		plan, "loop-registry-evidence",
	); chainSummary != "" {
		parts = append(parts, chainSummary)
	}
	if entrypointSummary := targetVerifierEntrypointSummaryFor(
		plan, "loop-registry-evidence",
	); entrypointSummary != "" {
		parts = append(parts, entrypointSummary)
	}
	if commandSummary := targetVerifierCommandSummary(plan); commandSummary != "" {
		parts = append(parts, commandSummary)
	}
	return strings.Join(parts, " ; ")
}
