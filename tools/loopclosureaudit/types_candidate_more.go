package main

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

type runRecord struct {
	ID      string `json:"id,omitempty"`
	Attempt string `json:"attempt,omitempty"`
	SHA     string `json:"sha,omitempty"`
	RefName string `json:"ref_name,omitempty"`
	Event   string `json:"event,omitempty"`
}
