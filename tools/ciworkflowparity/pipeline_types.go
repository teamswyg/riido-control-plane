package main

type pipeline struct {
	SchemaVersion string            `json:"schema_version"`
	ID            string            `json:"id"`
	Status        string            `json:"status"`
	Repo          string            `json:"repo"`
	Visibility    string            `json:"visibility"`
	Execution     pipelineExecution `json:"execution"`
	Evidence      pipelineEvidence  `json:"evidence_contract"`
	Steps         []pipelineStep    `json:"steps"`
	SuccessGate   []string          `json:"success_gate"`
}

type pipelineExecution struct {
	DefaultEngine string `json:"default_engine"`
	NativePolicy  string `json:"native_policy"`
	Attestation   string `json:"attestation"`
}

type pipelineEvidence struct {
	Artifact      string          `json:"artifact"`
	OwnerPackages []string        `json:"owner_packages"`
	Cases         []pipelineNamed `json:"cases"`
	SourceChecks  []pipelineNamed `json:"source_checks"`
	Loop          evidenceLoop    `json:"loop"`
}

type pipelineNamed struct {
	Name string `json:"name"`
}

type pipelineStep struct {
	ID        string   `json:"id"`
	Kind      string   `json:"kind"`
	Command   string   `json:"command"`
	Paths     []string `json:"paths"`
	Redaction string   `json:"redaction"`
	RunWhen   string   `json:"run_when"`
}
