package riidoaiserver

import "time"

const AIAgentClientContractID = "control-plane-ai-agent-client-api.v1"

const AgentInstructionMaxCharacters = 1000
const AgentDescriptionMaxCharacters = 160

const (
	AgentClientEventDeviceRuntimeSnapshot = "device_runtime_snapshot"
	AgentClientEventDeviceDaemonStatus    = "device_daemon_status_changed"
	AgentClientEventEditabilityChanged    = "agent_editability_changed"
	AgentClientEventWorkStatusChanged     = "agent_work_status_changed"
	AgentClientEventThreadProgress        = "agent_thread_progress"
)

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

type DaemonAvailability string

const (
	DaemonAvailabilityOnline  DaemonAvailability = "online"
	DaemonAvailabilityOffline DaemonAvailability = "offline"
)

type DaemonControlAction string

const (
	DaemonControlActionStart   DaemonControlAction = "start"
	DaemonControlActionRestart DaemonControlAction = "restart"
	DaemonControlActionStop    DaemonControlAction = "stop"
)

type DaemonControlState string

const (
	DaemonControlStateIdle       DaemonControlState = "idle"
	DaemonControlStateStarting   DaemonControlState = "starting"
	DaemonControlStateRestarting DaemonControlState = "restarting"
	DaemonControlStateStopping   DaemonControlState = "stopping"
	DaemonControlStateFailed     DaemonControlState = "failed"
)

type RuntimeModelRecord struct {
	ModelID   string `json:"model_id"`
	Label     string `json:"label"`
	IsDefault bool   `json:"is_default"`
}

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
	AgentTaskCommentAssignmentStarted     AgentTaskCommentKind = "assignment_started"
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
	Models           []RuntimeModelRecord  `json:"models"`
}

type DeviceRecord struct {
	DeviceID         string          `json:"device_id"`
	OwnerPrincipalID string          `json:"owner_principal_id"`
	DisplayName      string          `json:"display_name,omitempty"`
	DaemonLastSeenAt time.Time       `json:"daemon_last_seen_at,omitempty"`
	Runtimes         []RuntimeRecord `json:"runtimes"`
}

type DeviceDaemonRecord struct {
	DeviceID               string                `json:"device_id"`
	OwnerPrincipalID       string                `json:"owner_principal_id"`
	DeviceDisplayName      string                `json:"device_display_name,omitempty"`
	DaemonID               string                `json:"daemon_id,omitempty"`
	Profile                string                `json:"profile,omitempty"`
	PID                    int                   `json:"pid,omitempty"`
	UptimeSeconds          int                   `json:"uptime_seconds,omitempty"`
	StartedAt              time.Time             `json:"started_at,omitempty"`
	LastSeenAt             time.Time             `json:"last_seen_at,omitempty"`
	Availability           DaemonAvailability    `json:"availability"`
	ControlState           DaemonControlState    `json:"control_state"`
	SupportedActions       []DaemonControlAction `json:"supported_actions"`
	LastCommandID          string                `json:"last_command_id,omitempty"`
	LastCommandAction      DaemonControlAction   `json:"last_command_action,omitempty"`
	LastCommandRequestedAt time.Time             `json:"last_command_requested_at,omitempty"`
}

type DeviceDaemonDetailResponse struct {
	SchemaVersion string             `json:"schema_version"`
	Daemon        DeviceDaemonRecord `json:"daemon"`
}

type ControlDeviceDaemonRequest struct {
	Reason string `json:"reason,omitempty"`
}

type DeviceDaemonCommandResponse struct {
	SchemaVersion string              `json:"schema_version"`
	CommandID     string              `json:"command_id"`
	DeviceID      string              `json:"device_id"`
	Action        DaemonControlAction `json:"action"`
	Availability  DaemonAvailability  `json:"availability"`
	ControlState  DaemonControlState  `json:"control_state"`
	AcceptedAt    time.Time           `json:"accepted_at"`
	Message       string              `json:"message"`
}

type AgentClientRecord struct {
	AgentID             string           `json:"agent_id"`
	OwnerPrincipalID    string           `json:"owner_principal_id"`
	IsOwnedByViewer     bool             `json:"is_owned_by_viewer"`
	Name                string           `json:"name"`
	ProfileThumbnailURL string           `json:"profile_thumbnail_url,omitempty"`
	Description         string           `json:"description,omitempty"`
	Instruction         string           `json:"instruction,omitempty"`
	Visibility          AgentVisibility  `json:"visibility"`
	RuntimeID           string           `json:"runtime_id,omitempty"`
	RuntimeKind         RuntimeKind      `json:"runtime_kind,omitempty"`
	ModelID             string           `json:"model_id,omitempty"`
	ModelLabel          string           `json:"model_label,omitempty"`
	WorkStatus          AgentWorkStatus  `json:"work_status"`
	Editability         AgentEditability `json:"editability"`
	AssignedTaskCount   int              `json:"assigned_task_count"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

type AgentClientListResponse struct {
	SchemaVersion string              `json:"schema_version"`
	Agents        []AgentClientRecord `json:"agents"`
}

type AgentClientRecordResponse struct {
	SchemaVersion string            `json:"schema_version"`
	Agent         AgentClientRecord `json:"agent"`
}

type AgentOnboardingTemplate struct {
	TemplateID          string `json:"template_id"`
	Name                string `json:"name"`
	RoleLabel           string `json:"role_label,omitempty"`
	ProfileThumbnailURL string `json:"profile_thumbnail_url,omitempty"`
	Description         string `json:"description"`
	Instruction         string `json:"instruction"`
}

type ClientBootstrapResponse struct {
	SchemaVersion  string                    `json:"schema_version"`
	ClientKind     ClientKind                `json:"client_kind"`
	WorkspaceID    string                    `json:"workspace_id"`
	Agents         []AgentClientRecord       `json:"agents"`
	Devices        []DeviceRecord            `json:"devices"`
	AgentTemplates []AgentOnboardingTemplate `json:"agent_templates"`
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

type CreateAgentConfigurationRequest struct {
	Name                string          `json:"name"`
	ProfileThumbnailURL *string         `json:"profile_thumbnail_url,omitempty"`
	Description         *string         `json:"description,omitempty"`
	Instruction         *string         `json:"instruction,omitempty"`
	Visibility          AgentVisibility `json:"visibility"`
	RuntimeID           string          `json:"runtime_id"`
	ModelID             *string         `json:"model_id,omitempty"`
}

type UpdateAgentConfigurationRequest struct {
	Name                string          `json:"name,omitempty"`
	ProfileThumbnailURL *string         `json:"profile_thumbnail_url,omitempty"`
	Description         *string         `json:"description,omitempty"`
	Instruction         *string         `json:"instruction,omitempty"`
	Visibility          AgentVisibility `json:"visibility,omitempty"`
	RuntimeID           string          `json:"runtime_id,omitempty"`
	ModelID             *string         `json:"model_id,omitempty"`
}

type DeleteAgentResponse struct {
	SchemaVersion            string `json:"schema_version"`
	AgentID                  string `json:"agent_id"`
	QueuedTasksUnassigned    int    `json:"queued_tasks_unassigned"`
	RunningTasksForceStopped int    `json:"running_tasks_force_stopped"`
}

type AssignAIAgentTaskRequest struct {
	AgentID string `json:"agent_id"`
}

type UnassignAIAgentTaskRequest struct {
	AgentID string `json:"agent_id"`
	Reason  string `json:"reason,omitempty"`
}

type SubmitAIAgentTaskCommentRequest struct {
	AgentID         string `json:"agent_id"`
	Body            string `json:"body"`
	SourceCommentID string `json:"source_comment_id,omitempty"`
}

type CreateAIAgentTaskThreadMessageRequest struct {
	Body            string `json:"body"`
	SourceMessageID string `json:"source_message_id,omitempty"`
}

type StopAIAgentTaskRequest struct {
	AgentID string `json:"agent_id,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

type AIAgentTaskActionResponse struct {
	SchemaVersion   string               `json:"schema_version"`
	TaskID          string               `json:"task_id"`
	AgentID         string               `json:"agent_id"`
	ThreadID        string               `json:"thread_id"`
	RunID           string               `json:"run_id"`
	WorkStatus      AgentWorkStatus      `json:"work_status"`
	AssignmentState AgentAssignmentState `json:"assignment_state"`
	CommentKind     AgentTaskCommentKind `json:"comment_kind"`
	Message         string               `json:"message"`
}

type AIAgentTaskThreadStreamLink struct {
	Rel       string `json:"rel"`
	Href      string `json:"href"`
	EventType string `json:"event_type"`
	TaskID    string `json:"task_id"`
	ThreadID  string `json:"thread_id"`
	RunID     string `json:"run_id"`
}

type AIAgentTaskThreadRecord struct {
	ThreadID        string                    `json:"thread_id"`
	TaskID          string                    `json:"task_id"`
	AgentID         string                    `json:"agent_id"`
	RunID           string                    `json:"run_id"`
	SourceCommentID string                    `json:"source_comment_id,omitempty"`
	SourceMessageID string                    `json:"source_message_id,omitempty"`
	WorkStatus      AgentWorkStatus           `json:"work_status"`
	AssignmentState AgentAssignmentState      `json:"assignment_state"`
	CommentKind     AgentTaskCommentKind      `json:"comment_kind"`
	Message         string                    `json:"message"`
	StartedAt       time.Time                 `json:"started_at,omitempty"`
	CompletedAt     time.Time                 `json:"completed_at,omitempty"`
	Lines           []AgentThreadProgressLine `json:"lines"`
}

type AIAgentTaskThreadCollectionResponse struct {
	SchemaVersion string                       `json:"schema_version"`
	TaskID        string                       `json:"task_id"`
	Threads       []AIAgentTaskThreadRecord    `json:"threads"`
	ActiveStream  *AIAgentTaskThreadStreamLink `json:"active_stream,omitempty"`
}

type DeviceRuntimeSnapshotEvent struct {
	EventType     string       `json:"event_type"`
	SchemaVersion string       `json:"schema_version"`
	Device        DeviceRecord `json:"device"`
}

type DeviceDaemonStatusEvent struct {
	EventType     string             `json:"event_type"`
	SchemaVersion string             `json:"schema_version"`
	Daemon        DeviceDaemonRecord `json:"daemon"`
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
	ThreadID        string               `json:"thread_id,omitempty"`
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
	ThreadID        string                    `json:"thread_id"`
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
	ThreadID       string                    `json:"thread_id,omitempty"`
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
