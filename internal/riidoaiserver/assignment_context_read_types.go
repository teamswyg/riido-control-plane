package riidoaiserver

import "time"

// AssignmentContextRead records the requester, source, run and read time.
type AssignmentContextRead struct {
	AssignmentID         string    `json:"assignment_id"`
	TaskID               string    `json:"task_id"`
	AgentID              string    `json:"agent_id"`
	RequesterPrincipalID string    `json:"requester_principal_id,omitempty"`
	TriggerKind          string    `json:"trigger_kind"`
	TriggerID            string    `json:"trigger_id,omitempty"`
	SourceKind           string    `json:"source_kind"`
	SourceID             string    `json:"source_id,omitempty"`
	SourcePart           string    `json:"source_part"`
	SourceVersion        string    `json:"source_version,omitempty"`
	SourceUpdatedAt      string    `json:"source_updated_at,omitempty"`
	Via                  string    `json:"via"`
	Status               string    `json:"status"`
	ServedAt             time.Time `json:"served_at"`
	RequestID            string    `json:"request_id,omitempty"`
}
