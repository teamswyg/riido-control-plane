package main

func candidateSummaryForEvidence(
	m manifest,
	coverageGaps []claimCoverageGap,
) candidateSummary {
	return candidateSummary{
		CandidateArtifact: candidateArtifactForEvidence(m),
		CandidateCount:    candidateCountForEvidence(m, coverageGaps),
		CandidateIDs:      candidateIDsForEvidence(m, coverageGaps),
		CandidateSourceID: candidateSourceIDForEvidence(m),
		PromotionTarget:   candidateTargetForEvidence(m),
	}
}

func candidateIDsForEvidence(
	m manifest,
	coverageGaps []claimCoverageGap,
) []string {
	if len(m.Sources) == 0 {
		return nil
	}
	source := m.Sources[0]
	ids := make([]string, 0, len(m.ResidualGaps)+len(coverageGaps))
	for _, gap := range m.ResidualGaps {
		ids = append(ids, source.ID+":"+gap.ID)
	}
	for _, gap := range coverageGaps {
		ids = append(ids, claimCoverageCandidateID(source, gap))
	}
	return ids
}
