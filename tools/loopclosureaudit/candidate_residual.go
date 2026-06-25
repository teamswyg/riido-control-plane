package main

func residualGapCandidates(
	source candidateSource,
	gaps []residualGap,
	generatedAt string,
	expiresAt string,
) []closedLoopCandidate {
	out := make([]closedLoopCandidate, 0, len(gaps))
	for _, gap := range gaps {
		out = append(out, closedLoopCandidate{
			ID:                    source.ID + ":" + gap.ID,
			SourceRef:             candidateRef(source, generatedAt, expiresAt),
			HarnessLoop:           source.HarnessLoop,
			PromotionTarget:       source.PromotionTarget,
			PromotionEdge:         graphEdge{source.HarnessLoop, source.PromotionTarget, "promotes_failure_to"},
			Observation:           gap.Observation,
			Hypothesis:            gap.Risk,
			RequiredNextArtifacts: append([]string(nil), source.RequiredNextArtifacts...),
			AdoptionPlan:          adoptionPlan(source, gap),
		})
	}
	return out
}
