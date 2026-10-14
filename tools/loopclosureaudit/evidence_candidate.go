package main

func candidateCountForEvidence(
	m manifest,
	coverageGaps []claimCoverageGap,
) int {
	return len(m.ResidualGaps) + len(coverageGaps)
}

func candidateArtifactForEvidence(m manifest) string {
	if len(m.Sources) == 0 {
		return ""
	}
	return m.Sources[0].CandidateArtifact
}

func candidateSourceIDForEvidence(m manifest) string {
	if len(m.Sources) == 0 {
		return ""
	}
	return m.Sources[0].ID
}

func candidateTargetForEvidence(m manifest) string {
	if len(m.Sources) == 0 {
		return ""
	}
	return m.Sources[0].PromotionTarget
}
