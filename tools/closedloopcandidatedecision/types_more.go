package main

type verifyResult struct {
	DecisionCount     int                        `json:"decision_count"`
	CandidateCount    int                        `json:"candidate_count"`
	DecisionIDs       []string                   `json:"decision_ids"`
	DecisionArtifacts []decisionArtifactEvidence `json:"decision_artifacts"`
}

type decisionArtifactEvidence struct {
	CandidateID  string `json:"candidate_id"`
	NextArtifact string `json:"next_artifact"`
	NextCommand  string `json:"next_command"`
}

type candidateEvidence struct {
	SchemaVersion  string                `json:"schema_version"`
	ID             string                `json:"id"`
	Status         string                `json:"status"`
	LiveStatus     string                `json:"live_status"`
	CandidateCount int                   `json:"candidate_count"`
	Candidates     []closedLoopCandidate `json:"candidates"`
	Redaction      candidateRedaction    `json:"redaction"`
}

type closedLoopCandidate struct {
	ID                    string         `json:"id"`
	PromotionTarget       string         `json:"promotion_target"`
	Observation           string         `json:"observation"`
	Hypothesis            string         `json:"hypothesis"`
	RequiredNextArtifacts []string       `json:"required_next_artifacts"`
	AdoptionPlan          []adoptionStep `json:"adoption_plan"`
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
