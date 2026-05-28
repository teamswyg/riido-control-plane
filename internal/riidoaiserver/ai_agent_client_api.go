package riidoaiserver

import "time"

const AIAgentClientContractID = "control-plane-ai-agent-client-api.v1"

type AgentVisibility string

const (
	AgentVisibilityPublic  AgentVisibility = "public"
	AgentVisibilityPrivate AgentVisibility = "private"
)

type RuntimeKind string

const (
	RuntimeKindCodex      RuntimeKind = "codex"
	RuntimeKindClaudeCode RuntimeKind = "claude_code"
	RuntimeKindCursor     RuntimeKind = "cursor"
	RuntimeKindOpenClaw   RuntimeKind = "openclaw"
)

type RuntimeAvailability string

const (
	RuntimeAvailabilityOnline  RuntimeAvailability = "online"
	RuntimeAvailabilityOffline RuntimeAvailability = "offline"
)

type RuntimeDetectionState string

const (
	RuntimeDetectionStateDetected RuntimeDetectionState = "detected"
	RuntimeDetectionStateMissing  RuntimeDetectionState = "missing"
	RuntimeDetectionStateError    RuntimeDetectionState = "error"
)

type AgentEditability string

const (
	AgentEditabilityEditable             AgentEditability = "editable"
	AgentEditabilityBlockedAssignedTasks AgentEditability = "blocked_assigned_tasks"
)

type AgentWorkStatus string

const (
	AgentWorkStatusIdle           AgentWorkStatus = "idle"
	AgentWorkStatusQueued         AgentWorkStatus = "queued"
	AgentWorkStatusRunning        AgentWorkStatus = "running"
	AgentWorkStatusWaitingForUser AgentWorkStatus = "waiting_for_user"
	AgentWorkStatusCompleted      AgentWorkStatus = "completed"
	AgentWorkStatusFailed         AgentWorkStatus = "failed"
	AgentWorkStatusOffline        AgentWorkStatus = "offline"
)

type AgentAssignmentState string

const (
	AgentAssignmentStateQueued     AgentAssignmentState = "queued"
	AgentAssignmentStateRunning    AgentAssignmentState = "running"
	AgentAssignmentStateStopping   AgentAssignmentState = "stopping"
	AgentAssignmentStateStopped    AgentAssignmentState = "stopped"
	AgentAssignmentStateCompleted  AgentAssignmentState = "completed"
	AgentAssignmentStateFailed     AgentAssignmentState = "failed"
	AgentAssignmentStateUnassigned AgentAssignmentState = "unassigned"
)

type AgentTaskCommentKind string

const (
	AgentTaskCommentQueuedByBusyAgent     AgentTaskCommentKind = "queued_by_busy_agent"
	AgentTaskCommentStoppedByAgentDeleted AgentTaskCommentKind = "stopped_by_agent_deleted"
	AgentTaskCommentStoppedByUserRequest  AgentTaskCommentKind = "stopped_by_user_request"
	AgentTaskCommentRuntimeProgress       AgentTaskCommentKind = "runtime_progress"
	AgentTaskCommentTaskCompleted         AgentTaskCommentKind = "task_completed"
	AgentTaskCommentTaskFailed            AgentTaskCommentKind = "task_failed"
)

type ClientKind string

const (
	ClientKindWeb            ClientKind = "web"
	ClientKindDesktopWebview ClientKind = "desktop_webview"
)

type RuntimeRecord struct {
	RuntimeID        string                `json:"runtime_id"`
	DeviceID         string                `json:"device_id"`
	Kind             RuntimeKind           `json:"kind"`
	Availability     RuntimeAvailability   `json:"availability"`
	DetectionState   RuntimeDetectionState `json:"detection_state"`
	OwnerPrincipalID string                `json:"owner_principal_id,omitempty"`
	LastDetectedAt   time.Time             `json:"last_detected_at,omitempty"`
	HasAssignedAgent bool                  `json:"has_assigned_agent"`
}

type DeviceRecord struct {
	DeviceID         string          `json:"device_id"`
	OwnerPrincipalID string          `json:"owner_principal_id"`
	DisplayName      string          `json:"display_name,omitempty"`
	DaemonLastSeenAt time.Time       `json:"daemon_last_seen_at,omitempty"`
	Runtimes         []RuntimeRecord `json:"runtimes"`
}

type AgentClientRecord struct {
	AgentID           string           `json:"agent_id"`
	OwnerPrincipalID  string           `json:"owner_principal_id"`
	IsOwnedByViewer   bool             `json:"is_owned_by_viewer"`
	Name              string           `json:"name"`
	Visibility        AgentVisibility  `json:"visibility"`
	RuntimeID         string           `json:"runtime_id,omitempty"`
	RuntimeKind       RuntimeKind      `json:"runtime_kind,omitempty"`
	WorkStatus        AgentWorkStatus  `json:"work_status"`
	Editability       AgentEditability `json:"editability"`
	AssignedTaskCount int              `json:"assigned_task_count"`
}

type AgentClientListResponse struct {
	SchemaVersion string              `json:"schema_version"`
	Agents        []AgentClientRecord `json:"agents"`
}

type AgentClientRecordResponse struct {
	SchemaVersion string            `json:"schema_version"`
	Agent         AgentClientRecord `json:"agent"`
}

type ClientBootstrapResponse struct {
	SchemaVersion string              `json:"schema_version"`
	ClientKind    ClientKind          `json:"client_kind"`
	WorkspaceID   string              `json:"workspace_id"`
	Agents        []AgentClientRecord `json:"agents"`
	Devices       []DeviceRecord      `json:"devices"`
}

type DeviceRuntimeListResponse struct {
	SchemaVersion string         `json:"schema_version"`
	Devices       []DeviceRecord `json:"devices"`
}

type AgentEditabilityResponse struct {
	SchemaVersion     string           `json:"schema_version"`
	AgentID           string           `json:"agent_id"`
	Editability       AgentEditability `json:"editability"`
	AssignedTaskCount int              `json:"assigned_task_count"`
	Reason            string           `json:"reason,omitempty"`
}

type UpdateAgentConfigurationRequest struct {
	Name       string          `json:"name,omitempty"`
	Visibility AgentVisibility `json:"visibility,omitempty"`
	RuntimeID  string          `json:"runtime_id,omitempty"`
}

type DeleteAgentResponse struct {
	SchemaVersion            string `json:"schema_version"`
	AgentID                  string `json:"agent_id"`
	QueuedTasksUnassigned    int    `json:"queued_tasks_unassigned"`
	RunningTasksForceStopped int    `json:"running_tasks_force_stopped"`
}

type SubmitAIAgentTaskCommentRequest struct {
	AgentID         string `json:"agent_id"`
	Body            string `json:"body"`
	SourceCommentID string `json:"source_comment_id,omitempty"`
}

type StopAIAgentTaskRequest struct {
	AgentID string `json:"agent_id,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

type AIAgentTaskActionResponse struct {
	SchemaVersion   string               `json:"schema_version"`
	TaskID          string               `json:"task_id"`
	AgentID         string               `json:"agent_id"`
	RunID           string               `json:"run_id"`
	WorkStatus      AgentWorkStatus      `json:"work_status"`
	AssignmentState AgentAssignmentState `json:"assignment_state"`
	CommentKind     AgentTaskCommentKind `json:"comment_kind"`
	Message         string               `json:"message"`
}

type DeviceRuntimeSnapshotEvent struct {
	EventType     string       `json:"event_type"`
	SchemaVersion string       `json:"schema_version"`
	Device        DeviceRecord `json:"device"`
}

type AgentEditabilityChangedEvent struct {
	EventType         string           `json:"event_type"`
	SchemaVersion     string           `json:"schema_version"`
	AgentID           string           `json:"agent_id"`
	Editability       AgentEditability `json:"editability"`
	AssignedTaskCount int              `json:"assigned_task_count,omitempty"`
}

type AgentWorkStatusChangedEvent struct {
	EventType       string               `json:"event_type"`
	SchemaVersion   string               `json:"schema_version"`
	AgentID         string               `json:"agent_id"`
	TaskID          string               `json:"task_id,omitempty"`
	RunID           string               `json:"run_id,omitempty"`
	WorkStatus      AgentWorkStatus      `json:"work_status"`
	AssignmentState AgentAssignmentState `json:"assignment_state,omitempty"`
	CommentKind     AgentTaskCommentKind `json:"comment_kind,omitempty"`
}

type AgentThreadProgressLine struct {
	Seq        int       `json:"seq"`
	Message    string    `json:"message"`
	ObservedAt time.Time `json:"observed_at,omitempty"`
}

type AgentThreadProgressEvent struct {
	EventType       string                    `json:"event_type"`
	SchemaVersion   string                    `json:"schema_version"`
	AgentID         string                    `json:"agent_id"`
	TaskID          string                    `json:"task_id"`
	RunID           string                    `json:"run_id"`
	WorkStatus      AgentWorkStatus           `json:"work_status"`
	AssignmentState AgentAssignmentState      `json:"assignment_state"`
	CommentKind     AgentTaskCommentKind      `json:"comment_kind"`
	BatchStartedAt  time.Time                 `json:"batch_started_at,omitempty"`
	BatchEndedAt    time.Time                 `json:"batch_ended_at,omitempty"`
	Lines           []AgentThreadProgressLine `json:"lines"`
}

type AgentThreadProgressBatchRequest struct {
	AssignmentID   string                    `json:"assignment_id"`
	TaskID         string                    `json:"task_id"`
	DaemonID       string                    `json:"daemon_id,omitempty"`
	DeviceID       string                    `json:"device_id,omitempty"`
	RuntimeID      string                    `json:"runtime_id,omitempty"`
	RunID          string                    `json:"run_id,omitempty"`
	BatchStartedAt time.Time                 `json:"batch_started_at,omitempty"`
	BatchEndedAt   time.Time                 `json:"batch_ended_at,omitempty"`
	Lines          []AgentThreadProgressLine `json:"lines"`
	Metadata       map[string]string         `json:"metadata,omitempty"`
}

type AgentThreadProgressBatchResponse struct {
	SchemaVersion string                   `json:"schema_version"`
	AcceptedLines int                      `json:"accepted_lines"`
	Event         AgentThreadProgressEvent `json:"event"`
}

type ClientStreamEvent struct {
	Seq       int64
	EventType string
	Payload   any
}
