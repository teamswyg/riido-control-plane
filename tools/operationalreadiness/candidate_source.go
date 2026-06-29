package main

func readinessCandidateSource(m manifest) producerSource {
	for _, source := range m.Sources {
		if source.ID == readinessCandidateSourceID {
			return source
		}
	}
	return producerSource{
		ID:                    readinessCandidateSourceID,
		SourceWorkflow:        ".github/workflows/operational-readiness.yml",
		CandidateArtifact:     readinessCandidateArtifact,
		HarnessLoop:           readinessHarnessLoop,
		PromotionTarget:       readinessPromotionTarget,
		RequiredNextArtifacts: candidateRequiredArtifacts(),
	}
}

func readinessSourceRef(
	source producerSource,
	generatedAt string,
	expiresAt string,
	run runRecord,
) candidateSourceRef {
	return candidateSourceRef{
		HarnessLoop:       source.HarnessLoop,
		SourceWorkflow:    source.SourceWorkflow,
		SummaryArtifact:   "operational-readiness-evidence",
		CandidateArtifact: source.CandidateArtifact,
		LiveStatus:        "stale_partial",
		SourceGeneratedAt: generatedAt,
		SourceExpiresAt:   expiresAt,
		Run:               run,
	}
}

func candidateRequiredArtifacts() []string {
	return []string{
		"claim_binding",
		"verifier",
		"ci_gate",
		"redacted_evidence",
		"decision_record",
		"evidence_graph_edge",
	}
}
