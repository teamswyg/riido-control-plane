package main

const candidateSchema = "riido-control-plane-closed-loop-candidates.v1"

func newCandidateEvidence(m manifest) candidateEvidence {
	source := m.Sources[0]
	generatedAt, expiresAt := candidateWindow()
	candidates := residualGapCandidates(source, m.ResidualGaps, generatedAt, expiresAt)
	return candidateEvidence{
		SchemaVersion:     candidateSchema,
		ID:                source.ID,
		Status:            "verified",
		SourceWorkflow:    source.SourceWorkflow,
		LiveStatus:        "residual_gaps",
		SourceGeneratedAt: generatedAt,
		SourceExpiresAt:   expiresAt,
		CandidateCount:    len(candidates),
		Candidates:        candidates,
		Redaction:         candidateRedaction{true, true, true, true},
	}
}

func residualGapCandidates(source candidateSource, gaps []residualGap, generatedAt, expiresAt string) []closedLoopCandidate {
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
