package main

func staleSourceAdoptionPlan() []adoptionStep {
	steps := []adoptionStep{}
	for _, artifact := range requiredNextArtifacts() {
		steps = append(steps, adoptionStep{Artifact: artifact, Command: staleSourceAdoptionCommand(artifact)})
	}
	return steps
}

func staleSourceAdoptionCommand(artifact string) string {
	switch artifact {
	case "claim_binding":
		return "go run ./tools/loopregistry -check-doc -github-annotations"
	case "verifier":
		return "go test ./tools/looprefreshdispatch -run '^(TestLoopRefreshDispatchCLIWritesStaleSourceCandidate)$' -count=1"
	case "ci_gate":
		return "go test ./tools/looprefreshdispatch ./tools/loopregistry ./tools/evidencegraph -count=1"
	case "redacted_evidence":
		return "go run ./tools/looprefreshdispatch -commands-in out/fresh-loop-refresh-commands.json -evidence-out out/loop-refresh-dispatch-plan.json -candidate-out out/loop-refresh-dispatch-closed-loop-candidates.json"
	case "decision_record":
		return "go run ./tools/closedloopcandidatedecision -candidate-in out/loop-refresh-dispatch-closed-loop-candidates.json -check-doc"
	case "evidence_graph_edge":
		return "go run ./tools/evidencegraph -check-doc -evidence-out out/evidence-graph-evidence.json"
	default:
		return "unknown adoption artifact: " + artifact
	}
}
