package main

func targetVerifierSummaryParts(
	plan *targetVerifierPlan,
	evidenceOut string,
) []string {
	summaries := []string{
		targetVerifierPathSummaryFor(plan, evidenceOut),
		targetVerifierComponentSummaryFor(plan, evidenceOut),
		targetVerifierLoopSummaryFor(plan, evidenceOut),
		targetVerifierClaimSummaryFor(plan, evidenceOut),
		targetVerifierChainSummaryFor(plan, evidenceOut),
		targetVerifierRunnableSummaryFor(plan, evidenceOut),
		targetVerifierRouteSummaryFor(plan, evidenceOut),
		targetVerifierFocusedSummaryFor(plan, evidenceOut),
		targetVerifierEntrypointSummaryFor(plan, evidenceOut),
		targetVerifierCommandSummaryFor(plan, evidenceOut),
	}
	parts := []string{}
	for _, summary := range summaries {
		if summary != "" {
			parts = append(parts, "riido target verifier "+summary)
		}
	}
	return parts
}
