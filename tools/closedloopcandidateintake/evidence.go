package main

type evidence struct {
	SchemaVersion     string                      `json:"schema_version"`
	ID                string                      `json:"id"`
	Status            string                      `json:"status"`
	GeneratedAt       string                      `json:"generated_at"`
	ExpiresAt         string                      `json:"expires_at"`
	SourceCount       int                         `json:"source_count"`
	RequiredRefs      int                         `json:"required_refs"`
	CandidateCount    int                         `json:"candidate_count"`
	CandidateIDs      []string                    `json:"candidate_ids"`
	CandidateEdges    []graphEdge                 `json:"candidate_edges"`
	ConsumedArtifacts []consumedCandidateArtifact `json:"consumed_candidate_artifacts"`
	Workflow          string                      `json:"workflow"`
	GeneratedDoc      string                      `json:"generated_doc"`
	EvidenceArtifact  string                      `json:"evidence_artifact"`
	Loop              evidenceLoop                `json:"loop"`
	Sources           []intakeSource              `json:"sources"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	generatedAt, expiresAt := evidenceWindow(candidateIntakeEvidenceTTLHours)
	return evidence{
		SchemaVersion:     evidenceSchema,
		ID:                m.ID,
		Status:            "verified",
		GeneratedAt:       generatedAt,
		ExpiresAt:         expiresAt,
		SourceCount:       result.SourceCount,
		RequiredRefs:      result.RequiredRefs,
		CandidateCount:    result.CandidateCount,
		CandidateIDs:      result.CandidateIDs,
		CandidateEdges:    result.CandidateEdges,
		ConsumedArtifacts: result.ConsumedCandidateArtifacts,
		Workflow:          m.Workflow,
		GeneratedDoc:      m.GeneratedDoc,
		EvidenceArtifact:  m.EvidenceArtifact,
		Loop:              m.Loop,
		Sources:           m.Sources,
	}
}
