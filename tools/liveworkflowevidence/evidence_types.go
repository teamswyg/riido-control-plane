package main

type verifyResult struct {
	WorkflowCount int              `json:"workflow_count"`
	PhraseChecks  int              `json:"phrase_checks"`
	Records       []workflowRecord `json:"records"`
}

type workflowRecord struct {
	ID               string   `json:"id"`
	Path             string   `json:"path"`
	SummaryArtifact  string   `json:"summary_artifact"`
	SummaryPath      string   `json:"summary_path"`
	EvidenceTTLHours int      `json:"evidence_ttl_hours"`
	SensitiveInputs  []string `json:"sensitive_inputs"`
	RequiredPhrases  []string `json:"required_phrases,omitempty"`
	EvidenceClaims   []string `json:"evidence_claims,omitempty"`
}

type manifestEvidence struct {
	SchemaVersion string           `json:"schema_version"`
	ID            string           `json:"id"`
	Status        string           `json:"status"`
	GeneratedDoc  string           `json:"generated_doc"`
	Workflow      string           `json:"workflow"`
	WorkflowCount int              `json:"workflow_count"`
	PhraseChecks  int              `json:"phrase_checks"`
	Loop          loopRecord       `json:"loop"`
	Records       []workflowRecord `json:"records"`
}

type liveSummary struct {
	SchemaVersion    string             `json:"schema_version"`
	ID               string             `json:"id"`
	Status           string             `json:"status"`
	Workflow         workflowRecord     `json:"workflow"`
	Run              runRecord          `json:"run"`
	GeneratedAt      string             `json:"generated_at"`
	ExpiresAt        string             `json:"expires_at"`
	LiveStatus       string             `json:"live_status"`
	DeploymentTarget string             `json:"deployment_target,omitempty"`
	DeploymentMode   string             `json:"deployment_mode,omitempty"`
	BuildCacheMode   string             `json:"build_cache_mode,omitempty"`
	EvidenceClaims   []liveClaim        `json:"evidence_claims,omitempty"`
	Redaction        redactionAssertion `json:"redaction"`
}

type liveClaim struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
}

type runRecord struct {
	ID      string `json:"id,omitempty"`
	Attempt string `json:"attempt,omitempty"`
	SHA     string `json:"sha,omitempty"`
	RefName string `json:"ref_name,omitempty"`
	Event   string `json:"event,omitempty"`
}

type redactionAssertion struct {
	SummaryOnly        bool     `json:"summary_only"`
	NoRawSecrets       bool     `json:"no_raw_secrets"`
	NoRawEndpoints     bool     `json:"no_raw_endpoints"`
	NoAWSResourceIDs   bool     `json:"no_aws_resource_ids"`
	AllowedFields      []string `json:"allowed_fields"`
	SensitiveFieldRefs []string `json:"sensitive_field_refs"`
}
