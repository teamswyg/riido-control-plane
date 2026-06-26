package main

type refreshCommandEvidence struct {
	SchemaVersion string                   `json:"schema_version"`
	Status        string                   `json:"status"`
	GeneratedAt   string                   `json:"generated_at,omitempty"`
	ExpiresAt     string                   `json:"expires_at,omitempty"`
	CommandCount  int                      `json:"command_count"`
	Commands      []selectedRefreshCommand `json:"commands"`
}

type selectedRefreshCommand struct {
	LoopID  string `json:"loop_id"`
	Kind    string `json:"kind"`
	Command string `json:"command"`
}

type dispatchPlan struct {
	SchemaVersion       string             `json:"schema_version"`
	Status              string             `json:"status"`
	GeneratedAt         string             `json:"generated_at"`
	ExpiresAt           string             `json:"expires_at"`
	SourceStatus        string             `json:"source_status"`
	SourceGeneratedAt   string             `json:"source_generated_at,omitempty"`
	SourceExpiresAt     string             `json:"source_expires_at,omitempty"`
	DispatchCount       int                `json:"dispatch_count"`
	Dispatches          []workflowDispatch `json:"dispatches"`
	IgnoredCommandCount int                `json:"ignored_command_count"`
	IgnoredCommandKinds []string           `json:"ignored_command_kinds,omitempty"`
}

type workflowDispatch struct {
	WorkflowFile    string   `json:"workflow_file"`
	VerifiedCommand string   `json:"verified_command"`
	LoopIDs         []string `json:"loop_ids"`
	CommandCount    int      `json:"command_count"`
}
