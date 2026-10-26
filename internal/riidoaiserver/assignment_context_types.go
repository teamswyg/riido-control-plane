package riidoaiserver

import (
	"context"
	"time"
)

// Assignment context snapshots implement ADR-001 option 4 of the Riido CLI
// plan: the task context the server already fetched with the requester's
// Bearer at assign/reply time is kept per assignment, and the daemon reads it
// back with its device credential and lease token. No user token is stored.
const (
	AssignmentContextViaPrompt = "prompt"
	AssignmentContextViaCLI    = "cli"

	AssignmentContextReadServed      = "served"
	AssignmentContextReadTruncated   = "truncated"
	AssignmentContextReadDenied      = "denied"
	AssignmentContextReadUnavailable = "unavailable"

	AssignmentContextTriggerAssign        = "assign"
	AssignmentContextTriggerThreadMessage = "thread_message"
	AssignmentContextTriggerComment       = "comment"
)

type AssignmentContextSnapshot struct {
	AssignmentID         string             `json:"assignment_id"`
	TaskID               string             `json:"task_id"`
	WorkspaceID          string             `json:"workspace_id,omitempty"`
	AgentID              string             `json:"agent_id"`
	RequesterPrincipalID string             `json:"requester_principal_id,omitempty"`
	TriggerKind          string             `json:"trigger_kind"`
	TriggerID            string             `json:"trigger_id,omitempty"`
	CapturedAt           time.Time          `json:"captured_at"`
	Truncated            bool               `json:"truncated,omitempty"`
	Context              AIAgentTaskContext `json:"context"`
}

// AssignmentContextRead is one "source read for this answer" record: who
// (agent and requester), what (source kind/id/part/version), when, and which
// run (assignment plus the triggering comment or thread message).
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
	Via                  string    `json:"via"`
	Status               string    `json:"status"`
	ServedAt             time.Time `json:"served_at"`
	RequestID            string    `json:"request_id,omitempty"`
}

type AssignmentContextStore interface {
	SaveAssignmentContextSnapshot(ctx context.Context, snapshot AssignmentContextSnapshot) error
	LoadAssignmentContextSnapshot(ctx context.Context, assignmentID string) (AssignmentContextSnapshot, bool, error)
	AppendAssignmentContextReads(ctx context.Context, reads []AssignmentContextRead) error
	ListAssignmentContextReads(ctx context.Context, assignmentID string) ([]AssignmentContextRead, error)
}
