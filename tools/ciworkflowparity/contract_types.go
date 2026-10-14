package main

type manifest struct {
	SchemaVersion   string         `json:"schema_version"`
	ID              string         `json:"id"`
	Status          string         `json:"status"`
	Issue           string         `json:"issue"`
	ParentIssue     string         `json:"parent_issue"`
	CheckedOn       string         `json:"checked_on"`
	Pipelines       []string       `json:"pipelines"`
	Baseline        baseline       `json:"baseline"`
	BoundedChildren []boundedChild `json:"bounded_children"`
	Runner          runner         `json:"runner"`
	NativeMapping   nativeMapping  `json:"native_mapping"`
	ParityClaim     parityClaim    `json:"parity_claim"`
	Authority       authority      `json:"authority"`
	Rollback        rollback       `json:"rollback"`
	Classification  classification `json:"classification"`
	Assertions      []string       `json:"assertions"`
	Loop            evidenceLoop   `json:"loop"`
}

type boundedChild struct {
	ID             string             `json:"id"`
	Issue          string             `json:"issue"`
	ParentIssue    string             `json:"parent_issue"`
	Baseline       baseline           `json:"baseline"`
	NativeMapping  childNativeMapping `json:"native_mapping"`
	ParityClaim    childParityClaim   `json:"parity_claim"`
	Authority      authority          `json:"authority"`
	Rollback       rollback           `json:"rollback"`
	Classification classification     `json:"classification"`
	Assertions     []string           `json:"assertions"`
	Loop           evidenceLoop       `json:"loop"`
}

type baseline struct {
	SourceRevision       string `json:"source_revision"`
	Workflow             string `json:"workflow"`
	WorkflowSHA256       string `json:"workflow_sha256"`
	WorkflowName         string `json:"workflow_name"`
	Job                  string `json:"job"`
	TrackedWorkflowCount int    `json:"tracked_workflow_count"`
}

type runner struct {
	Provider                 string `json:"provider"`
	Revision                 string `json:"revision"`
	Pipeline                 string `json:"pipeline"`
	PipelineID               string `json:"pipeline_id"`
	Visibility               string `json:"visibility"`
	GitHubTokenRequired      bool   `json:"github_token_required"`
	CloudCredentialsRequired bool   `json:"cloud_credentials_required"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
