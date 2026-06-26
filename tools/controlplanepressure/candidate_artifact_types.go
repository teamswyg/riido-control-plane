package main

type pressureCandidateEvidence struct {
	SchemaVersion     string                  `json:"schema_version"`
	ID                string                  `json:"id"`
	Status            string                  `json:"status"`
	SourceWorkflow    string                  `json:"source_workflow"`
	LiveStatus        string                  `json:"live_status"`
	SourceGeneratedAt string                  `json:"source_generated_at"`
	SourceExpiresAt   string                  `json:"source_expires_at"`
	Run               pressureRunRecord       `json:"run"`
	CandidateCount    int                     `json:"candidate_count"`
	Candidates        []pressureLoopCandidate `json:"candidates"`
	Redaction         pressureCandidateRedact `json:"redaction"`
}

type pressureLoopCandidate struct {
	ID                    string                     `json:"id"`
	SourceRef             pressureCandidateSourceRef `json:"source_ref"`
	HarnessLoop           string                     `json:"harness_loop"`
	PromotionTarget       string                     `json:"promotion_target"`
	PromotionEdge         pressureGraphEdge          `json:"promotion_edge"`
	Observation           string                     `json:"observation"`
	Hypothesis            string                     `json:"hypothesis"`
	RequiredNextArtifacts []string                   `json:"required_next_artifacts"`
	AdoptionPlan          []adoptionStep             `json:"adoption_plan"`
}
