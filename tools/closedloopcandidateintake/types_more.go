package main

type verifyResult struct {
	SourceCount                int                         `json:"source_count"`
	RequiredRefs               int                         `json:"required_refs"`
	CandidateCount             int                         `json:"candidate_count"`
	CandidateIDs               []string                    `json:"candidate_ids"`
	CandidateEdges             []graphEdge                 `json:"candidate_edges"`
	ConsumedCandidateArtifacts []consumedCandidateArtifact `json:"consumed_candidate_artifacts"`
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
	SummaryOnly      bool `json:"summary_only"`
	NoRawSecrets     bool `json:"no_raw_secrets"`
	NoRawEndpoints   bool `json:"no_raw_endpoints"`
	NoRawAWSResource bool `json:"no_raw_aws_resource_ids"`
}

type adoptionStep struct {
	Artifact string `json:"artifact"`
	Command  string `json:"command"`
}

type consumedCandidateArtifact struct {
	InputPath         string   `json:"input_path"`
	SourceWorkflow    string   `json:"source_workflow"`
	LiveStatus        string   `json:"live_status"`
	SourceGeneratedAt string   `json:"source_generated_at"`
	SourceExpiresAt   string   `json:"source_expires_at"`
	CandidateCount    int      `json:"candidate_count"`
	CandidateIDs      []string `json:"candidate_ids"`
	SourceIDs         []string `json:"source_ids"`
}
