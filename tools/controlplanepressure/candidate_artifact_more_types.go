package main

type pressureCandidateSourceRef struct {
	HarnessLoop       string            `json:"harness_loop"`
	SourceWorkflow    string            `json:"source_workflow"`
	SummaryArtifact   string            `json:"summary_artifact"`
	CandidateArtifact string            `json:"candidate_artifact"`
	LiveStatus        string            `json:"live_status"`
	SourceGeneratedAt string            `json:"source_generated_at"`
	SourceExpiresAt   string            `json:"source_expires_at"`
	Run               pressureRunRecord `json:"run"`
}

type pressureRunRecord struct {
	ID      string `json:"id,omitempty"`
	Attempt string `json:"attempt,omitempty"`
	SHA     string `json:"sha,omitempty"`
	RefName string `json:"ref_name,omitempty"`
	Event   string `json:"event,omitempty"`
}

type pressureGraphEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

type pressureCandidateRedact struct {
	SummaryOnly      bool `json:"summary_only"`
	NoRawSecrets     bool `json:"no_raw_secrets"`
	NoRawEndpoints   bool `json:"no_raw_endpoints"`
	NoRawAWSResource bool `json:"no_raw_aws_resource_ids"`
}
