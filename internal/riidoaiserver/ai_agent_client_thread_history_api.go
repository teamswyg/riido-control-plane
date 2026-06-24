package riidoaiserver

import "time"

type AIAgentTaskThreadMessageRole string

const (
	AIAgentTaskThreadMessageRoleUser     AIAgentTaskThreadMessageRole = "user"
	AIAgentTaskThreadMessageRoleAgent    AIAgentTaskThreadMessageRole = "agent"
	AIAgentTaskThreadMessageRoleProgress AIAgentTaskThreadMessageRole = "progress"
)

type AIAgentTaskThreadHistoryMessage struct {
	MessageID       string                       `json:"message_id"`
	Role            AIAgentTaskThreadMessageRole `json:"role"`
	CommentKind     AgentTaskCommentKind         `json:"comment_kind,omitempty"`
	AssignmentID    string                       `json:"assignment_id,omitempty"`
	RunID           string                       `json:"run_id,omitempty"`
	SourceMessageID string                       `json:"source_message_id,omitempty"`
	Seq             int                          `json:"seq,omitempty"`
	Body            string                       `json:"body,omitempty"`
	ResultMessage   string                       `json:"result_message,omitempty"`
	ObservedAt      time.Time                    `json:"observed_at,omitempty"`
}

type AIAgentTaskThreadHistoryRecord struct {
	ThreadID        string                            `json:"thread_id"`
	ConversationID  string                            `json:"conversation_id"`
	ParentThreadID  string                            `json:"parent_thread_id,omitempty"`
	TaskID          string                            `json:"task_id"`
	AssignmentID    string                            `json:"assignment_id,omitempty"`
	AgentID         string                            `json:"agent_id"`
	AgentSnapshotID string                            `json:"agent_snapshot_id,omitempty"`
	RunID           string                            `json:"run_id"`
	WorkStatus      AgentWorkStatus                   `json:"work_status"`
	AssignmentState AgentAssignmentState              `json:"assignment_state"`
	StartedAt       time.Time                         `json:"started_at,omitempty"`
	CompletedAt     time.Time                         `json:"completed_at,omitempty"`
	Messages        []AIAgentTaskThreadHistoryMessage `json:"messages"`
	ActiveStream    *AIAgentTaskThreadStreamLink      `json:"active_stream,omitempty"`
}

type AIAgentTaskThreadHistoryCollectionResponse struct {
	SchemaVersion  string                                    `json:"schema_version"`
	TaskID         string                                    `json:"task_id"`
	Threads        []AIAgentTaskThreadHistoryRecord          `json:"threads"`
	AgentSnapshots map[string]AIAgentTaskThreadAgentSnapshot `json:"agent_snapshots,omitempty"`
	ActiveStream   *AIAgentTaskThreadStreamLink              `json:"active_stream,omitempty"`
}
