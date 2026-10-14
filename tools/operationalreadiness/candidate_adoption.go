package main

func readinessAdoptionPlan(source producerSource) []adoptionStep {
	steps := []adoptionStep{}
	for _, artifact := range source.RequiredNextArtifacts {
		steps = append(steps, adoptionStep{Artifact: artifact, Command: readinessAdoptionCommand(artifact)})
	}
	return steps
}

func readinessAdoptionCommand(artifact string) string {
	switch artifact {
	case "claim_binding":
		return "go run ./tools/loopregistry -check-doc -github-annotations"
	case "verifier":
		return "go test ./tools/operationalreadiness ./tools/closedloopcandidateintake -count=1"
	case "ci_gate":
		return "go run ./tools/operationalreadiness -check-doc -evidence-out out/operational-readiness-evidence.json -candidate-out out/operational-readiness-closed-loop-candidates.json"
	case "redacted_evidence":
		return "go run ./tools/operationalreadiness -candidate-out out/operational-readiness-closed-loop-candidates.json"
	case "decision_record":
		return "go run ./tools/closedloopcandidatedecision -candidate-in out/operational-readiness-closed-loop-candidates.json -check-doc"
	case "evidence_graph_edge":
		return "go run ./tools/evidencegraph -check-doc -evidence-out out/evidence-graph-evidence.json"
	default:
		return "unknown adoption artifact: " + artifact
	}
}
