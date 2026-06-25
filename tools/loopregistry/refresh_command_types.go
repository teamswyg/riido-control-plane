package main

type refreshCommandEvidence struct {
	SchemaVersion     string                   `json:"schema_version"`
	Status            string                   `json:"status"`
	GeneratedAt       string                   `json:"generated_at"`
	SourceGeneratedAt string                   `json:"source_generated_at"`
	SourceExpiresAt   string                   `json:"source_expires_at"`
	SelectedLoopCount int                      `json:"selected_loop_count"`
	CommandCount      int                      `json:"command_count"`
	SelectedLoops     []selectedRefreshLoop    `json:"selected_loops"`
	Commands          []selectedRefreshCommand `json:"commands"`
}

type selectedRefreshLoop struct {
	LoopID            string `json:"loop_id"`
	EvidenceExpiresAt string `json:"evidence_expires_at"`
	CommandCount      int    `json:"command_count"`
}

type selectedRefreshCommand struct {
	LoopID  string `json:"loop_id"`
	Kind    string `json:"kind"`
	Command string `json:"command"`
}
