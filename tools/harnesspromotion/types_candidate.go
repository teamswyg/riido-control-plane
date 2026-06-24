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
	ID                    string   `json:"id"`
	HarnessLoop           string   `json:"harness_loop"`
	PromotionTarget       string   `json:"promotion_target"`
	Observation           string   `json:"observation"`
	Hypothesis            string   `json:"hypothesis"`
	RequiredNextArtifacts []string `json:"required_next_artifacts"`
}

type candidateRedaction struct {
	SummaryOnly      bool `json:"summary_only"`
	NoRawSecrets     bool `json:"no_raw_secrets"`
	NoRawEndpoints   bool `json:"no_raw_endpoints"`
	NoRawAWSResource bool `json:"no_raw_aws_resource_ids"`
}
