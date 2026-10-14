package main

type evidence struct {
	SchemaVersion           string            `json:"schema_version"`
	Decision                string            `json:"decision"`
	BaselineWorkflowSHA256  string            `json:"baseline_workflow_sha256"`
	PipelineID              string            `json:"pipeline_id"`
	RunnerRevision          string            `json:"runner_revision"`
	RequiredAdapterCount    int               `json:"required_adapter_count"`
	LegacyWorkflowPreserved bool              `json:"legacy_workflow_preserved"`
	RetirementAuthorized    bool              `json:"retirement_authorized"`
	RuntimeEffect           string            `json:"runtime_effect"`
	BoundedChildren         []childEvidence   `json:"bounded_children"`
	Cases                   map[string]string `json:"cases"`
}

type childEvidence struct {
	ID                      string `json:"id"`
	Issue                   string `json:"issue"`
	WorkflowSHA256          string `json:"workflow_sha256"`
	RequiredAdapterCount    int    `json:"required_adapter_count"`
	LegacyWorkflowPreserved bool   `json:"legacy_workflow_preserved"`
	RetirementAuthorized    bool   `json:"retirement_authorized"`
	RuntimeEffect           string `json:"runtime_effect"`
}
