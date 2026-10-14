package main

func staleSourceDecisionAdoptionPlan() []adoptionStep {
	return []adoptionStep{
		{"claim_binding", "go run ./tools/loopregistry -check-doc -github-annotations"},
		{"verifier", staleSourceVerifierCommand()},
		{"ci_gate", "go test ./tools/looprefreshdispatch ./tools/loopregistry ./tools/evidencegraph -count=1"},
		{"redacted_evidence", staleSourceEvidenceCommand()},
		{"decision_record", "go run ./tools/closedloopcandidatedecision -candidate-in out/loop-refresh-dispatch-closed-loop-candidates.json -check-doc"},
		{"evidence_graph_edge", "go run ./tools/evidencegraph -check-doc -evidence-out out/evidence-graph-evidence.json"},
	}
}

func staleSourceVerifierCommand() string {
	return "go test ./tools/looprefreshdispatch -run '^(TestLoopRefreshDispatchCLIWritesStaleSourceCandidate)$' -count=1"
}

func staleSourceEvidenceCommand() string {
	return "go run ./tools/looprefreshdispatch -commands-in out/fresh-loop-refresh-commands.json -evidence-out out/loop-refresh-dispatch-plan.json -candidate-out out/loop-refresh-dispatch-closed-loop-candidates.json"
}
