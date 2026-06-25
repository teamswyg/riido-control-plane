package main

type evidence struct {
	SchemaVersion    string                `json:"schema_version"`
	Status           string                `json:"status"`
	RequirementCount int                   `json:"requirement_count"`
	CheckCount       int                   `json:"check_count"`
	ResidualGapCount int                   `json:"residual_gap_count"`
	Requirements     []requirementEvidence `json:"requirements"`
	Assertions       []string              `json:"assertions"`
	ResidualGaps     []residualGap         `json:"residual_gaps"`
	Loop             loopSpec              `json:"loop"`
}

type requirementEvidence struct {
	ID         string   `json:"id"`
	Statement  string   `json:"statement"`
	CheckKinds []string `json:"check_kinds"`
	Checks     []check  `json:"checks"`
}

func newEvidence(m manifest) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		Status:           "verified",
		RequirementCount: len(m.Requirements),
		CheckCount:       checkCount(m.Requirements),
		ResidualGapCount: len(m.ResidualGaps),
		Requirements:     requirementEvidenceRows(m.Requirements),
		Assertions:       append([]string(nil), m.Assertions...),
		ResidualGaps:     append([]residualGap(nil), m.ResidualGaps...),
		Loop:             m.Loop,
	}
}

func checkCount(requirements []requirement) int {
	count := 0
	for _, req := range requirements {
		count += len(req.Checks)
	}
	return count
}

func requirementEvidenceRows(requirements []requirement) []requirementEvidence {
	rows := make([]requirementEvidence, 0, len(requirements))
	for _, req := range requirements {
		kinds := make([]string, 0, len(req.Checks))
		for _, check := range req.Checks {
			kinds = append(kinds, check.Kind)
		}
		rows = append(rows, requirementEvidence{
			ID:         req.ID,
			Statement:  req.Statement,
			CheckKinds: kinds,
			Checks:     evidenceChecks(req.Checks),
		})
	}
	return rows
}
