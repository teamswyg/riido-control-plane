package main

func pressureCandidate(
	candidate candidateEntry,
	measured pressureCandidateMeasurement,
	generatedAt string,
	expiresAt string,
) pressureLoopCandidate {
	return pressureLoopCandidate{
		ID: candidate.ID, SourceRef: pressureSourceRef(generatedAt, expiresAt),
		HarnessLoop: candidate.HarnessLoop, PromotionTarget: candidate.PromotionTarget,
		PromotionEdge:         pressureGraphEdge{candidate.HarnessLoop, candidate.PromotionTarget, "promotes_failure_to"},
		Observation:           "Local pressure harness measured " + candidate.Scenario + " as an optimization candidate.",
		Hypothesis:            candidate.Risk + " Next: " + candidate.Next,
		Measured:              measured,
		RequiredNextArtifacts: append([]string(nil), candidate.RequiredNextArtifacts...),
		AdoptionPlan:          append([]adoptionStep(nil), candidate.AdoptionPlan...),
	}
}

func pressureSourceRef(generatedAt, expiresAt string) pressureCandidateSourceRef {
	return pressureCandidateSourceRef{
		HarnessLoop: pressureHarnessLoop, SourceWorkflow: pressureSourceWorkflow,
		SummaryArtifact: pressureSummaryArtifact, CandidateArtifact: pressureCandidateArtifact,
		LiveStatus: pressureLiveStatus, SourceGeneratedAt: generatedAt,
		SourceExpiresAt: expiresAt, Run: githubRunRecord(),
	}
}
