package main

type evidence struct {
	SchemaVersion       string                       `json:"schema_version"`
	ID                  string                       `json:"id"`
	Status              string                       `json:"status"`
	GeneratedAt         string                       `json:"generated_at"`
	ExpiresAt           string                       `json:"expires_at"`
	DecisionCount       int                          `json:"decision_count"`
	CandidateCount      int                          `json:"candidate_count"`
	DecisionSummary     decisionSummary              `json:"decision_summary"`
	DecisionIDs         []string                     `json:"decision_ids"`
	DecisionArtifacts   []decisionArtifactEvidence   `json:"decision_artifacts"`
	CandidateSourceRefs []candidateSourceRefEvidence `json:"candidate_source_refs"`
	CandidateSubjects   []candidateSubjectEvidence   `json:"candidate_subjects"`
	ConsumedArtifacts   []consumedCandidateArtifact  `json:"consumed_candidate_artifacts"`
	Workflow            string                       `json:"workflow"`
	GeneratedDoc        string                       `json:"generated_doc"`
	EvidenceArtifact    string                       `json:"evidence_artifact"`
	CommandArtifact     string                       `json:"command_artifact"`
	Assertions          []string                     `json:"assertions"`
	Loop                evidenceLoop                 `json:"loop"`
	Decisions           []decisionRecord             `json:"decisions"`
	DecisionTemplates   []decisionTemplate           `json:"decision_templates,omitempty"`
}

func newEvidence(m manifest, result verifyResult) evidence {
	generatedAt, expiresAt := evidenceWindow(candidateDecisionEvidenceTTLHours)
	return evidence{
		SchemaVersion:       evidenceSchema,
		ID:                  m.ID,
		Status:              "verified",
		GeneratedAt:         generatedAt,
		ExpiresAt:           expiresAt,
		DecisionCount:       result.DecisionCount,
		CandidateCount:      result.CandidateCount,
		DecisionSummary:     newDecisionSummary(m, result),
		DecisionIDs:         result.DecisionIDs,
		DecisionArtifacts:   result.DecisionArtifacts,
		CandidateSourceRefs: result.CandidateSourceRefs,
		CandidateSubjects:   result.CandidateSubjects,
		ConsumedArtifacts:   result.ConsumedCandidateArtifacts,
		Workflow:            m.Workflow,
		GeneratedDoc:        m.GeneratedDoc,
		EvidenceArtifact:    m.EvidenceArtifact,
		CommandArtifact:     m.CommandArtifact,
		Assertions:          m.Assertions,
		Loop:                m.Loop,
		Decisions:           m.Decisions,
		DecisionTemplates:   m.DecisionTemplates,
	}
}
