package main

type evidence struct {
	SchemaVersion    string         `json:"schema_version"`
	ID               string         `json:"id"`
	Status           string         `json:"status"`
	SourceCount      int            `json:"source_count"`
	RequiredRefs     int            `json:"required_refs"`
	CandidateCount   int            `json:"candidate_count"`
	CandidateIDs     []string       `json:"candidate_ids"`
	Workflow         string         `json:"workflow"`
	GeneratedDoc     string         `json:"generated_doc"`
	EvidenceArtifact string         `json:"evidence_artifact"`
	Loop             evidenceLoop   `json:"loop"`
	Sources          []intakeSource `json:"sources"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	return evidence{
		SchemaVersion:    evidenceSchema,
		ID:               m.ID,
		Status:           "verified",
		SourceCount:      result.SourceCount,
		RequiredRefs:     result.RequiredRefs,
		CandidateCount:   result.CandidateCount,
		CandidateIDs:     result.CandidateIDs,
		Workflow:         m.Workflow,
		GeneratedDoc:     m.GeneratedDoc,
		EvidenceArtifact: m.EvidenceArtifact,
		Loop:             m.Loop,
		Sources:          m.Sources,
	}
}
