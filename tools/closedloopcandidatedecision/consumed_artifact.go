package main

func consumedArtifact(
	path string,
	candidate candidateEvidence,
	ids []string,
) consumedCandidateArtifact {
	return consumedCandidateArtifact{
		InputPath:         path,
		SourceWorkflow:    candidate.SourceWorkflow,
		LiveStatus:        candidate.LiveStatus,
		SourceGeneratedAt: candidate.SourceGeneratedAt,
		SourceExpiresAt:   candidate.SourceExpiresAt,
		CandidateCount:    candidate.CandidateCount,
		CandidateIDs:      append([]string(nil), ids...),
	}
}
