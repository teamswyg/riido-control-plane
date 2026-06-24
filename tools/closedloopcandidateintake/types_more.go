package main

type verifyResult struct {
	SourceCount    int      `json:"source_count"`
	RequiredRefs   int      `json:"required_refs"`
	CandidateCount int      `json:"candidate_count"`
	CandidateIDs   []string `json:"candidate_ids"`
}

type candidateEvidence struct {
	SchemaVersion  string                `json:"schema_version"`
	ID             string                `json:"id"`
	Status         string                `json:"status"`
	SourceWorkflow string                `json:"source_workflow"`
	LiveStatus     string                `json:"live_status"`
	CandidateCount int                   `json:"candidate_count"`
	Candidates     []closedLoopCandidate `json:"candidates"`
	Redaction      candidateRedaction    `json:"redaction"`
}

type closedLoopCandidate struct {
	ID                    string         `json:"id"`
	HarnessLoop           string         `json:"harness_loop"`
	PromotionTarget       string         `json:"promotion_target"`
	Observation           string         `json:"observation"`
	Hypothesis            string         `json:"hypothesis"`
	RequiredNextArtifacts []string       `json:"required_next_artifacts"`
	AdoptionPlan          []adoptionStep `json:"adoption_plan"`
}

type candidateRedaction struct {
	SummaryOnly      bool `json:"summary_only"`
	NoRawSecrets     bool `json:"no_raw_secrets"`
	NoRawEndpoints   bool `json:"no_raw_endpoints"`
	NoRawAWSResource bool `json:"no_raw_aws_resource_ids"`
}

type adoptionStep struct {
	Artifact string `json:"artifact"`
	Command  string `json:"command"`
}
