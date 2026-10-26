package riidoaiserver

import "time"

type AgentAssignmentContextRequest struct {
	AssignmentID string `json:"assignment_id"`
	LeaseToken   string `json:"lease_token"`
	View         string `json:"view"`
	Target       string `json:"target,omitempty"`
	Cursor       string `json:"cursor,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
}

type AgentAssignmentContextResponse struct {
	SchemaVersion string                    `json:"schema_version"`
	AssignmentID  string                    `json:"assignment_id"`
	TaskID        string                    `json:"task_id"`
	View          string                    `json:"view"`
	CapturedAt    time.Time                 `json:"captured_at"`
	Truncated     bool                      `json:"truncated,omitempty"`
	Data          any                       `json:"data"`
	Sources       []AssignmentContextSource `json:"sources"`
}
