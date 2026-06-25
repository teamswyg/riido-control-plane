package main

func claimCoverageHypothesis(gap claimCoverageGap) string {
	return "Agents can find files and tests for " + gap.ClaimID +
		", but cannot infer its loop observation, verifier, or failure semantics from machine tokens alone."
}

func claimCoverageAdoptionPlan(source candidateSource, gap claimCoverageGap) []adoptionStep {
	out := make([]adoptionStep, 0, len(source.RequiredNextArtifacts))
	for _, artifact := range source.RequiredNextArtifacts {
		out = append(out, adoptionStep{Artifact: artifact, Command: claimCoverageCommand(artifact, gap)})
	}
	return out
}

func claimCoverageCommand(artifact string, gap claimCoverageGap) string {
	switch artifact {
	case "decision_record":
		return "go run ./tools/closedloopcandidatedecision -check-doc -candidate-in out/loop-closure-audit-closed-loop-candidates.json"
	case "evidence_graph_edge":
		return "go run ./tools/evidencegraph -check-doc -evidence-out out/evidence-graph.json"
	default:
		return "go run ./tools/loopregistry -check-doc -evidence-out out/loop-registry-evidence.json"
	}
}
