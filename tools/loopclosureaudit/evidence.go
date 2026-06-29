package main

type evidence struct {
	SchemaVersion         string                `json:"schema_version"`
	Status                string                `json:"status"`
	GeneratedAt           string                `json:"generated_at"`
	ExpiresAt             string                `json:"expires_at"`
	RequirementCount      int                   `json:"requirement_count"`
	CheckCount            int                   `json:"check_count"`
	ResidualGapCount      int                   `json:"residual_gap_count"`
	ClaimCoverageGapCount int                   `json:"claim_coverage_gap_count"`
	CandidateCount        int                   `json:"candidate_count"`
	CandidateArtifact     string                `json:"candidate_artifact,omitempty"`
	CandidateSourceID     string                `json:"candidate_source_id,omitempty"`
	CandidateTarget       string                `json:"candidate_promotion_target,omitempty"`
	Requirements          []requirementEvidence `json:"requirements"`
	Assertions            []string              `json:"assertions"`
	ResidualGaps          []residualGap         `json:"residual_gaps"`
	ClaimCoverageGaps     []claimCoverageGap    `json:"claim_coverage_gaps,omitempty"`
	Loop                  loopSpec              `json:"loop"`
}

type requirementEvidence struct {
	ID         string   `json:"id"`
	Statement  string   `json:"statement"`
	CheckKinds []string `json:"check_kinds"`
	Checks     []check  `json:"checks"`
}

type claimCoverageGap struct {
	ClaimID           string   `json:"claim_id"`
	Loop              string   `json:"loop"`
	MissingDimensions []string `json:"missing_dimensions"`
}

func newEvidence(m manifest, depsOpt ...dependencies) evidence {
	coverageGaps := []claimCoverageGap{}
	if len(depsOpt) > 0 {
		coverageGaps = claimCoverageGaps(depsOpt[0])
	}
	generatedAt, expiresAt := evidenceWindow(loopClosureAuditEvidenceTTLHours)
	return evidence{
		SchemaVersion:         evidenceSchema,
		Status:                "verified",
		GeneratedAt:           generatedAt,
		ExpiresAt:             expiresAt,
		RequirementCount:      len(m.Requirements),
		CheckCount:            checkCount(m.Requirements),
		ResidualGapCount:      len(m.ResidualGaps),
		ClaimCoverageGapCount: len(coverageGaps),
		CandidateCount:        candidateCountForEvidence(m, coverageGaps),
		CandidateArtifact:     candidateArtifactForEvidence(m),
		CandidateSourceID:     candidateSourceIDForEvidence(m),
		CandidateTarget:       candidateTargetForEvidence(m),
		Requirements:          requirementEvidenceRows(m.Requirements),
		Assertions:            append([]string(nil), m.Assertions...),
		ResidualGaps:          append([]residualGap(nil), m.ResidualGaps...),
		ClaimCoverageGaps:     coverageGaps,
		Loop:                  m.Loop,
	}
}
