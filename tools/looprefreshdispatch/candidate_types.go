package main

type candidateEvidence struct {
	SchemaVersion     string                `json:"schema_version"`
	ID                string                `json:"id"`
	Status            string                `json:"status"`
	SourceWorkflow    string                `json:"source_workflow"`
	LiveStatus        string                `json:"live_status"`
	SourceGeneratedAt string                `json:"source_generated_at"`
	SourceExpiresAt   string                `json:"source_expires_at"`
	Run               runRecord             `json:"run"`
	CandidateCount    int                   `json:"candidate_count"`
	Candidates        []closedLoopCandidate `json:"candidates"`
	Redaction         candidateRedaction    `json:"redaction"`
}

type closedLoopCandidate struct {
	ID                    string             `json:"id"`
	SourceRef             candidateSourceRef `json:"source_ref"`
	HarnessLoop           string             `json:"harness_loop"`
	PromotionTarget       string             `json:"promotion_target"`
	PromotionEdge         graphEdge          `json:"promotion_edge"`
	Observation           string             `json:"observation"`
	Hypothesis            string             `json:"hypothesis"`
	RequiredNextArtifacts []string           `json:"required_next_artifacts"`
	AdoptionPlan          []adoptionStep     `json:"adoption_plan"`
}

type candidateSourceRef struct {
	HarnessLoop       string    `json:"harness_loop"`
	SourceWorkflow    string    `json:"source_workflow"`
	SummaryArtifact   string    `json:"summary_artifact"`
	CandidateArtifact string    `json:"candidate_artifact"`
	LiveStatus        string    `json:"live_status"`
	SourceGeneratedAt string    `json:"source_generated_at"`
	SourceExpiresAt   string    `json:"source_expires_at"`
	Run               runRecord `json:"run"`
}

type runRecord struct {
	ID      string `json:"id,omitempty"`
	Attempt string `json:"attempt,omitempty"`
	SHA     string `json:"sha,omitempty"`
	RefName string `json:"ref_name,omitempty"`
	Event   string `json:"event,omitempty"`
}
