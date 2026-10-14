package main

func testCandidateSource() producerSource {
	return producerSource{
		ID:                    readinessCandidateSourceID,
		SourceWorkflow:        ".github/workflows/operational-readiness.yml",
		CandidateArtifact:     readinessCandidateArtifact,
		HarnessLoop:           readinessHarnessLoop,
		PromotionTarget:       readinessPromotionTarget,
		RequiredNextArtifacts: candidateRequiredArtifacts(),
	}
}
