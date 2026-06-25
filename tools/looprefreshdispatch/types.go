package main

type refreshCommandEvidence struct {
	SchemaVersion string                   `json:"schema_version"`
	Status        string                   `json:"status"`
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
	SourceStatus        string             `json:"source_status"`
	DispatchCount       int                `json:"dispatch_count"`
	Dispatches          []workflowDispatch `json:"dispatches"`
	IgnoredCommandCount int                `json:"ignored_command_count"`
	IgnoredCommandKinds []string           `json:"ignored_command_kinds,omitempty"`
}

type workflowDispatch struct {
	WorkflowFile string   `json:"workflow_file"`
	LoopIDs      []string `json:"loop_ids"`
	CommandCount int      `json:"command_count"`
}
