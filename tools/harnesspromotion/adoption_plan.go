package main

func adoptionPlan(source promotionSource) []adoptionStep {
	steps := []adoptionStep{}
	for _, artifact := range source.RequiredNextArtifacts {
		steps = append(steps, adoptionStep{
			Artifact: artifact,
			Command:  adoptionCommand(source, artifact),
		})
	}
	return steps
}

func adoptionCommand(source promotionSource, artifact string) string {
	switch artifact {
	case "claim_binding":
		return "go run ./tools/loopregistry -check-doc -github-annotations"
	case "verifier":
		return "go test ./tools/loopregistry ./tools/evidencegraph ./tools/harnesspromotion ./tools/closedloopcandidateintake ./tools/closedloopcandidatedecision -count=1"
	case "ci_gate":
		return "go run ./tools/loopregistry -check-doc -evidence-out out/loop-registry-evidence.json"
	case "redacted_evidence":
		return "go run ./tools/harnesspromotion -summary " + source.SummaryPath + " -candidate-out " + source.CandidatePath
	case "decision_record":
		return "go run ./tools/closedloopcandidatedecision -candidate-in " + source.CandidatePath + " -check-doc"
	case "evidence_graph_edge":
		return "go run ./tools/evidencegraph -check-doc -evidence-out out/evidence-graph-evidence.json"
	default:
		return "unknown adoption artifact: " + artifact
	}
}
