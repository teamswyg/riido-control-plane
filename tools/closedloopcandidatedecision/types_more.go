package main

type verifyResult struct {
	DecisionCount              int                          `json:"decision_count"`
	CandidateCount             int                          `json:"candidate_count"`
	DecisionIDs                []string                     `json:"decision_ids"`
	DecisionArtifacts          []decisionArtifactEvidence   `json:"decision_artifacts"`
	CandidateSourceRefs        []candidateSourceRefEvidence `json:"candidate_source_refs"`
	CandidateSubjects          []candidateSubjectEvidence   `json:"candidate_subjects"`
	ConsumedCandidateArtifacts []consumedCandidateArtifact  `json:"consumed_candidate_artifacts"`
}

type decisionArtifactEvidence struct {
	CandidateID                 string    `json:"candidate_id"`
	Disposition                 string    `json:"disposition"`
	Priority                    string    `json:"priority"`
	Owner                       string    `json:"owner"`
	ReviewBy                    string    `json:"review_by,omitempty"`
	NextLoop                    string    `json:"next_loop"`
	NextArtifact                string    `json:"next_artifact"`
	NextCommand                 string    `json:"next_command"`
	DecisionSource              string    `json:"decision_source"`
	DecisionTemplateSubjectKind string    `json:"decision_template_subject_kind,omitempty"`
	PromotionEdge               graphEdge `json:"promotion_edge"`
}

type candidateEvidence struct {
	SchemaVersion     string                `json:"schema_version"`
	ID                string                `json:"id"`
	Status            string                `json:"status"`
	SourceWorkflow    string                `json:"source_workflow"`
	LiveStatus        string                `json:"live_status"`
	SourceGeneratedAt string                `json:"source_generated_at"`
	SourceExpiresAt   string                `json:"source_expires_at"`
	CandidateCount    int                   `json:"candidate_count"`
	Candidates        []closedLoopCandidate `json:"candidates"`
	Redaction         candidateRedaction    `json:"redaction"`
}

type closedLoopCandidate struct {
	ID                    string             `json:"id"`
	SourceRef             candidateSourceRef `json:"source_ref"`
	Subject               rawSubject         `json:"subject,omitempty"`
	HarnessLoop           string             `json:"harness_loop"`
	PromotionTarget       string             `json:"promotion_target"`
	PromotionEdge         graphEdge          `json:"promotion_edge"`
	Observation           string             `json:"observation"`
	Hypothesis            string             `json:"hypothesis"`
	RequiredNextArtifacts []string           `json:"required_next_artifacts"`
	AdoptionPlan          []adoptionStep     `json:"adoption_plan"`
}

type graphEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

type candidateRedaction struct {
	SummaryOnly    bool `json:"summary_only"`
	NoRawSecrets   bool `json:"no_raw_secrets"`
	NoRawEndpoints bool `json:"no_raw_endpoints"`
}

type adoptionStep struct {
	Artifact string `json:"artifact"`
	Command  string `json:"command"`
}
