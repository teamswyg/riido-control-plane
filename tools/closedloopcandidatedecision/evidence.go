package main

type evidence struct {
	SchemaVersion    string           `json:"schema_version"`
	ID               string           `json:"id"`
	Status           string           `json:"status"`
	DecisionCount    int              `json:"decision_count"`
	CandidateCount   int              `json:"candidate_count"`
	DecisionIDs      []string         `json:"decision_ids"`
	Workflow         string           `json:"workflow"`
	GeneratedDoc     string           `json:"generated_doc"`
	EvidenceArtifact string           `json:"evidence_artifact"`
	Assertions       []string         `json:"assertions"`
	Loop             evidenceLoop     `json:"loop"`
	Decisions        []decisionRecord `json:"decisions"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		DecisionCount:    result.DecisionCount,
		CandidateCount:   result.CandidateCount,
		DecisionIDs:      result.DecisionIDs,
		Workflow:         m.Workflow,
		GeneratedDoc:     m.GeneratedDoc,
		EvidenceArtifact: m.EvidenceArtifact,
		Assertions:       m.Assertions,
		Loop:             m.Loop,
		Decisions:        m.Decisions,
	}
}
