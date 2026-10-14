package main

const loopCompletionThresholdBasisPoints = 9000

func loopCompletions(m manifest) []loopCompletion {
	claims := loopClaimCoverage(m.Claims)
	graph := loopGraphCoverage(m.EvidenceGraph)
	out := make([]loopCompletion, 0, len(m.Loops))
	for _, loop := range m.Loops {
		checks := loopCompletionChecks(loop, claims[loop.ID], graph[loop.ID])
		out = append(out, newLoopCompletion(loop, checks))
	}
	return out
}

func newLoopCompletion(loop loopRecord, checks map[string]bool) loopCompletion {
	missing := missingCompletionChecks(checks)
	passed := len(checks) - len(missing)
	bps := 0
	if len(checks) > 0 {
		bps = passed * 10000 / len(checks)
	}
	status := "verified"
	if bps < loopCompletionThresholdBasisPoints {
		status = "partial"
	}
	return loopCompletion{
		LoopID: loop.ID, Kind: loop.Kind, Status: status,
		RequiredChecks: len(checks), PassedChecks: passed,
		CompletionBasisPoints: bps,
		ThresholdBasisPoints:  loopCompletionThresholdBasisPoints,
		MissingChecks:         missing,
	}
}

func loopCompletionChecks(loop loopRecord, claims loopClaimSet, graph bool) map[string]bool {
	checks := map[string]bool{
		"observes_declared":        len(loop.Observes) > 0,
		"verifies_declared":        len(loop.Verifies) > 0,
		"fails_when_declared":      len(loop.FailsWhen) > 0,
		"evidence_declared":        len(loop.Evidence) > 0,
		"refresh_workflow":         loop.RefreshWorkflow != "",
		"expiry_declared":          loop.ExpiresAfterHours > 0,
		"claim_binding":            claims.Count > 0,
		"observes_claim_covered":   claims.CoversAll(loop.Observes, "observes"),
		"verifies_claim_covered":   claims.CoversAll(loop.Verifies, "verifies"),
		"fails_claim_covered":      claims.CoversAll(loop.FailsWhen, "fails"),
		"evidence_claim_covered":   claims.CoversEvidence(loop.Evidence),
		"evidence_graph_connected": graph,
	}
	if loop.Kind == "harness" {
		checks["harness_promotes_failure"] = len(loop.PromotesTo) > 0
	}
	return checks
}
