package main

func candidateRef(source candidateSource, generatedAt, expiresAt string) candidateSourceRef {
	return candidateSourceRef{
		HarnessLoop:       source.HarnessLoop,
		SourceWorkflow:    source.SourceWorkflow,
		SummaryArtifact:   source.SummaryArtifact,
		CandidateArtifact: source.CandidateArtifact,
		LiveStatus:        candidateLiveStatus,
		SourceGeneratedAt: generatedAt,
		SourceExpiresAt:   expiresAt,
		Run:               newCandidateRun(),
	}
}

func adoptionPlan(source candidateSource, gap residualGap) []adoptionStep {
	out := make([]adoptionStep, 0, len(source.RequiredNextArtifacts))
	for _, artifact := range source.RequiredNextArtifacts {
		out = append(out, adoptionStep{Artifact: artifact, Command: commandForArtifact(artifact, gap)})
	}
	return out
}

func commandForArtifact(artifact string, gap residualGap) string {
	switch artifact {
	case gap.NextArtifact:
		return gap.NextCommand
	case "decision_record":
		return "go run ./tools/closedloopcandidatedecision -check-doc -candidate-in out/loop-closure-audit-closed-loop-candidates.json"
	case "evidence_graph_edge":
		return "go run ./tools/evidencegraph -check-doc -evidence-out out/evidence-graph.json"
	default:
		return "go run ./tools/loopregistry -check-doc -evidence-out out/loop-registry-evidence.json"
	}
}
