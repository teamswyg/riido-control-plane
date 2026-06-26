package main

func requiredNextArtifacts() []string {
	return []string{
		"claim_binding",
		"verifier",
		"ci_gate",
		"redacted_evidence",
		"decision_record",
		"evidence_graph_edge",
	}
}

func adoptionPlan(command selectedRefreshCommand) []adoptionStep {
	steps := []adoptionStep{}
	for _, artifact := range requiredNextArtifacts() {
		steps = append(steps, adoptionStep{Artifact: artifact, Command: adoptionCommand(command, artifact)})
	}
	return steps
}

func adoptionCommand(command selectedRefreshCommand, artifact string) string {
	switch artifact {
	case "claim_binding":
		return "go run ./tools/loopregistry -check-doc -github-annotations"
	case "verifier":
		return command.Command
	case "ci_gate":
		return "go run ./tools/closedloopcandidateintake -check-doc -evidence-out out/closed-loop-candidate-intake-evidence.json"
	case "redacted_evidence":
		return "go run ./tools/looprefreshdispatch -commands-in out/loop-refresh-commands.json -evidence-out out/loop-refresh-dispatch-plan.json -candidate-out out/loop-refresh-dispatch-closed-loop-candidates.json"
	case "decision_record":
		return "go run ./tools/closedloopcandidatedecision -candidate-in out/loop-refresh-dispatch-closed-loop-candidates.json -check-doc"
	case "evidence_graph_edge":
		return "go run ./tools/evidencegraph -check-doc -evidence-out out/evidence-graph-evidence.json"
	default:
		return "unknown adoption artifact: " + artifact
	}
}
