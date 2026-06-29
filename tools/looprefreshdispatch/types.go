package main

type refreshCommandEvidence struct {
	SchemaVersion string                   `json:"schema_version"`
	Status        string                   `json:"status"`
	GeneratedAt   string                   `json:"generated_at,omitempty"`
	ExpiresAt     string                   `json:"expires_at,omitempty"`
	CommandCount  int                      `json:"command_count"`
	Commands      []selectedRefreshCommand `json:"commands"`
	SourcePath    string                   `json:"-"`
	StaleSources  []staleRefreshSource     `json:"stale_sources,omitempty"`
}

type selectedRefreshCommand struct {
	LoopID           string   `json:"loop_id"`
	Kind             string   `json:"kind"`
	Command          string   `json:"command"`
	CandidateID      string   `json:"candidate_id,omitempty"`
	SubjectKind      string   `json:"subject_kind,omitempty"`
	ClaimIDs         []string `json:"claim_ids,omitempty"`
	EvidenceChainIDs []string `json:"evidence_chain_ids,omitempty"`
}

type staleRefreshSource struct {
	SourcePath  string `json:"source_path,omitempty"`
	GeneratedAt string `json:"generated_at,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`
	Reason      string `json:"reason"`
}

type dispatchPlan struct {
	SchemaVersion       string                   `json:"schema_version"`
	Status              string                   `json:"status"`
	GeneratedAt         string                   `json:"generated_at"`
	ExpiresAt           string                   `json:"expires_at"`
	SourceStatus        string                   `json:"source_status"`
	SourceGeneratedAt   string                   `json:"source_generated_at,omitempty"`
	SourceExpiresAt     string                   `json:"source_expires_at,omitempty"`
	SourceCommandCount  int                      `json:"source_command_count"`
	SourceStaleCount    int                      `json:"source_stale_count,omitempty"`
	SourceStaleSources  []staleRefreshSource     `json:"source_stale_sources,omitempty"`
	DispatchCount       int                      `json:"dispatch_count"`
	Dispatches          []workflowDispatch       `json:"dispatches"`
	IgnoredCommandCount int                      `json:"ignored_command_count"`
	IgnoredCommandKinds []string                 `json:"ignored_command_kinds,omitempty"`
	IgnoredCommands     []selectedRefreshCommand `json:"ignored_commands,omitempty"`
}

type workflowDispatch struct {
	WorkflowFile    string   `json:"workflow_file"`
	VerifiedCommand string   `json:"verified_command"`
	LoopIDs         []string `json:"loop_ids"`
	CommandCount    int      `json:"command_count"`
}
