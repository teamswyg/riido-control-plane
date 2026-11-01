package riidoaiserver

import (
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
	// AssignmentContextTriggerAgentMention marks a run started by another
	// agent's mention (RIID-5892); its snapshot is inherited, not captured.
	AssignmentContextTriggerAgentMention = "agent_mention"
)

type AssignmentContextSnapshot struct {
	AssignmentID         string                      `json:"assignment_id"`
	ConversationID       string                      `json:"conversation_id,omitempty"`
	ThreadID             string                      `json:"thread_id,omitempty"`
	PreviousAssignmentID string                      `json:"previous_assignment_id,omitempty"`
	PreviousThreadID     string                      `json:"previous_thread_id,omitempty"`
	TaskID               string                      `json:"task_id"`
	WorkspaceID          string                      `json:"workspace_id,omitempty"`
	AgentID              string                      `json:"agent_id"`
	RequesterPrincipalID string                      `json:"requester_principal_id,omitempty"`
	RequesterAdmin       bool                        `json:"requester_admin,omitempty"`
	TriggerKind          string                      `json:"trigger_kind"`
	TriggerID            string                      `json:"trigger_id,omitempty"`
	CapturedAt           time.Time                   `json:"captured_at"`
	Truncated            bool                        `json:"truncated,omitempty"`
	CommentPages         int                         `json:"comment_pages,omitempty"`
	Content              []AssignmentContentManifest `json:"content,omitempty"`
	ContentBlocks        []AssignmentContentBlockRef `json:"content_blocks,omitempty"`
	Context              AIAgentTaskContext          `json:"context"`
}
