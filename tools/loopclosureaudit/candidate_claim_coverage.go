package main

import "strings"

func claimCoverageCandidates(
	source candidateSource,
	deps dependencies,
	generatedAt string,
	expiresAt string,
) []closedLoopCandidate {
	gaps := claimCoverageGaps(deps)
	out := make([]closedLoopCandidate, 0, len(gaps))
	for _, gap := range gaps {
		out = append(out, closedLoopCandidate{
			ID:                    claimCoverageCandidateID(source, gap),
			SourceRef:             candidateRef(source, generatedAt, expiresAt),
			Subject:               claimCoverageSubject(gap),
			HarnessLoop:           source.HarnessLoop,
			PromotionTarget:       source.PromotionTarget,
			PromotionEdge:         graphEdge{source.HarnessLoop, source.PromotionTarget, "promotes_failure_to"},
			Observation:           claimCoverageObservation(gap),
			Hypothesis:            claimCoverageHypothesis(gap),
			RequiredNextArtifacts: append([]string(nil), source.RequiredNextArtifacts...),
			AdoptionPlan:          claimCoverageAdoptionPlan(source, gap),
		})
	}
	return out
}

func claimCoverageCandidateID(source candidateSource, gap claimCoverageGap) string {
	return source.ID + ":claim_coverage:" + gap.ClaimID
}

func claimCoverageSubject(gap claimCoverageGap) *candidateSubject {
	return &candidateSubject{
		Kind:              "claim_coverage_gap",
		ClaimID:           gap.ClaimID,
		Loop:              gap.Loop,
		MissingDimensions: append([]string(nil), gap.MissingDimensions...),
	}
}

func claimCoverageObservation(gap claimCoverageGap) string {
	return "Claim " + gap.ClaimID + " in loop " + gap.Loop +
		" is missing coverage dimensions: " + strings.Join(gap.MissingDimensions, ", ") + "."
}
