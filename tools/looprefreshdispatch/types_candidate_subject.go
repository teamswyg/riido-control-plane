package main

type candidateSubject struct {
	Kind        string `json:"kind"`
	LoopID      string `json:"loop_id,omitempty"`
	CommandKind string `json:"command_kind,omitempty"`
	Command     string `json:"command,omitempty"`
	SourcePath  string `json:"source_path,omitempty"`
	GeneratedAt string `json:"generated_at,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`
	Reason      string `json:"reason,omitempty"`
}
