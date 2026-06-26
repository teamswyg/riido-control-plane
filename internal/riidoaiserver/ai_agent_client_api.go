package riidoaiserver

import (
	"encoding/json"
	"time"
)

const (
	AIAgentClientContractID               = "control-plane-ai-agent-client-api.v2"
	AIAgentClientPersistenceSchemaVersion = "riido-ai-agent-client-persistence.v2"
)

const (
	AgentInstructionMaxCharacters          = 1000
	AgentDescriptionMaxCharacters          = 160
	AIAgentDeviceRuntimeSnapshotStaleAfter = 20 * time.Second
)

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
	RuntimeID                 string                `json:"runtime_id"`
	DeviceID                  string                `json:"device_id"`
	DaemonID                  string                `json:"daemon_id,omitempty"`
	DaemonProfile             string                `json:"daemon_profile,omitempty"`
	Kind                      RuntimeKind           `json:"kind"`
	Availability              RuntimeAvailability   `json:"availability"`
	DetectionState            RuntimeDetectionState `json:"detection_state"`
	ProviderVersion           string                `json:"provider_version"`
	OwnerPrincipalID          string                `json:"owner_principal_id,omitempty"`
	LastDetectedAt            time.Time             `json:"last_detected_at,omitempty"`
	HasAssignedAgent          bool                  `json:"has_assigned_agent"`
	RequiresExperimentalOptIn bool                  `json:"requires_experimental_opt_in"`
	Models                    []RuntimeModelRecord  `json:"models"`
}

type RuntimeSnapshotRecord struct {
	RuntimeID                 string                `json:"runtime_id"`
	Kind                      RuntimeKind           `json:"kind"`
	Availability              RuntimeAvailability   `json:"availability,omitempty"`
	DetectionState            RuntimeDetectionState `json:"detection_state,omitempty"`
	ProviderVersion           string                `json:"provider_version,omitempty"`
	RequiresExperimentalOptIn bool                  `json:"requires_experimental_opt_in,omitempty"`
	Models                    []RuntimeModelRecord  `json:"models,omitempty"`
}

type DeviceRecord struct {
	DeviceID         string    `json:"device_id"`
	OwnerPrincipalID string    `json:"owner_principal_id"`
	DisplayName      string    `json:"display_name,omitempty"`
	DaemonLastSeenAt time.Time `json:"daemon_last_seen_at,omitempty"`
	// ConnectedWorkspaceIDs is the set of workspaces this (machine) device is
	// connected to. The device is visible to every member of any workspace it is
	// connected to — visibility is workspace-connection scoped, not owner scoped.
	// A device connects to a workspace when it enrolls or reports a runtime
	// snapshot in that workspace.
	ConnectedWorkspaceIDs []string        `json:"connected_workspace_ids,omitempty"`
	Runtimes              []RuntimeRecord `json:"runtimes"`
}

type DeviceRuntimeSnapshotSyncRequest struct {
	DaemonID          string                  `json:"daemon_id"`
	DeviceID          string                  `json:"device_id,omitempty"`
	DeviceDisplayName string                  `json:"device_display_name,omitempty"`
	Profile           string                  `json:"profile,omitempty"`
	AppVersion        string                  `json:"app_version,omitempty"`
	PID               int                     `json:"pid,omitempty"`
	UptimeSeconds     int                     `json:"uptime_seconds,omitempty"`
	StartedAt         time.Time               `json:"started_at,omitempty"`
	Runtimes          []RuntimeSnapshotRecord `json:"runtimes"`
}

type DeviceRuntimeSnapshotSyncResponse struct {
	SchemaVersion string             `json:"schema_version"`
	Device        DeviceRecord       `json:"device"`
	Daemon        DeviceDaemonRecord `json:"daemon"`
}

type AgentRuntimeBindingListResponse struct {
	SchemaVersion string                `json:"schema_version"`
	Bindings      []AgentRuntimeBinding `json:"bindings"`
}

type DeviceDaemonRecord struct {
	DeviceID               string                `json:"device_id"`
	OwnerPrincipalID       string                `json:"owner_principal_id"`
	DeviceDisplayName      string                `json:"device_display_name,omitempty"`
	DaemonID               string                `json:"daemon_id,omitempty"`
	Profile                string                `json:"profile,omitempty"`
	AppVersion             string                `json:"app_version,omitempty"`
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
	Runtime       *RuntimeRecord     `json:"runtime,omitempty"`
	Runtimes      []RuntimeRecord    `json:"runtimes,omitempty"`
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
	WorkspaceID         string           `json:"workspace_id,omitempty"`
	IsOwnedByViewer     bool             `json:"is_owned_by_viewer"`
	Name                string           `json:"name"`
	ProfileThumbnailURL string           `json:"profile_thumbnail_url,omitempty"`
	TmpColor            string           `json:"tmp_color,omitempty"`
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

type AgentOnboardingFixture struct {
	FixtureID              string          `json:"fixture_id"`
	Name                   string          `json:"name"`
	RoleLabel              string          `json:"role_label,omitempty"`
	ProfileThumbnailURL    string          `json:"profile_thumbnail_url,omitempty"`
	TmpColor               string          `json:"tmp_color,omitempty"`
	Description            string          `json:"description"`
	Instruction            string          `json:"instruction"`
	DefaultVisibility      AgentVisibility `json:"default_visibility"`
	RecommendedRuntimeKind RuntimeKind     `json:"recommended_runtime_kind,omitempty"`
}

type AgentOnboardingFixtureListResponse struct {
	SchemaVersion string                   `json:"schema_version"`
	Fixtures      []AgentOnboardingFixture `json:"fixtures"`
}

type AssignedAgentProfile struct {
	AvatarURL string `json:"avatar_url,omitempty"`
	TmpColor  string `json:"tmp_color,omitempty"`
}

type AssignedAgentProfileMapResponse struct {
	SchemaVersion         string                          `json:"schema_version"`
	WorkspaceID           string                          `json:"workspace_id"`
	AssignedAgentProfiles map[string]AssignedAgentProfile `json:"assigned_agent_profiles"`
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

type AgentProfileThumbnailUploadFormField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CreateAgentProfileThumbnailUploadRequest struct {
	ContentType        string `json:"content_type"`
	ContentLengthBytes int64  `json:"content_length_bytes"`
	FileName           string `json:"file_name,omitempty"`
}

type AgentProfileThumbnailUploadResponse struct {
	SchemaVersion         string                                 `json:"schema_version"`
	Method                string                                 `json:"method"`
	UploadURL             string                                 `json:"upload_url"`
	FormFileField         string                                 `json:"form_file_field"`
	FormFields            []AgentProfileThumbnailUploadFormField `json:"form_fields"`
	ProfileThumbnailURL   string                                 `json:"profile_thumbnail_url"`
	ExpiresAt             time.Time                              `json:"expires_at"`
	MaxContentLengthBytes int64                                  `json:"max_content_length_bytes"`
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
	AgentID            string `json:"agent_id"`
	AssignmentID       string `json:"-"`
	durableState       AssignmentState
	intentGateRequired bool
}

type UnassignAIAgentTaskRequest struct {
	AgentID      string `json:"agent_id"`
	AssignmentID string `json:"assignment_id,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

type AgentAssignmentActionRequest struct {
	AssignmentID string `json:"assignment_id,omitempty"`
	Reason       string `json:"reason,omitempty"`
	durableState AssignmentState
}

type SubmitAIAgentTaskCommentRequest struct {
	AgentID         string `json:"agent_id"`
	Body            string `json:"body"`
	SourceCommentID string `json:"source_comment_id,omitempty"`
}

type CreateAIAgentTaskThreadMessageRequest struct {
	Body            string `json:"body"`
	SourceMessageID string `json:"source_message_id,omitempty"`
	AssignmentID    string `json:"-"`
	toolApproval    bool
	durableState    AssignmentState
}

type StopAIAgentTaskRequest struct {
	AgentID      string `json:"agent_id,omitempty"`
	AssignmentID string `json:"assignment_id,omitempty"`
	Reason       string `json:"reason,omitempty"`
	durableState AssignmentState
}

type AIAgentTaskActionResponse struct {
	SchemaVersion      string                               `json:"schema_version"`
	TaskID             string                               `json:"task_id"`
	AssignmentID       string                               `json:"assignment_id,omitempty"`
	AgentID            string                               `json:"agent_id"`
	AgentSnapshot      *AIAgentTaskThreadAgentSnapshot      `json:"agent_snapshot,omitempty"`
	ThreadID           string                               `json:"thread_id"`
	RunID              string                               `json:"run_id"`
	WorkStatus         AgentWorkStatus                      `json:"work_status"`
	AssignmentState    AgentAssignmentState                 `json:"assignment_state"`
	CommentKind        AgentTaskCommentKind                 `json:"comment_kind"`
	Message            string                               `json:"message"`
	ResultMessage      string                               `json:"result_message,omitempty"`
	FailureDiagnostics *AIAgentTaskThreadFailureDiagnostics `json:"failure_diagnostics,omitempty"`
	ActiveStream       *AIAgentTaskThreadStreamLink         `json:"active_stream,omitempty"`
}

type AIAgentTaskThreadStreamLink struct {
	Rel       string `json:"rel"`
	Href      string `json:"href"`
	EventType string `json:"event_type"`
	TaskID    string `json:"task_id"`
	ThreadID  string `json:"thread_id"`
	RunID     string `json:"run_id"`
}

type AIAgentTaskEventStreamLink struct {
	Rel       string `json:"rel"`
	Href      string `json:"href"`
	EventType string `json:"event_type"`
}

type AIAgentTaskThreadStreamTarget struct {
	AgentID  string `json:"agent_id"`
	ThreadID string `json:"thread_id"`
	RunID    string `json:"run_id"`
}

type AIAgentTaskThreadQueueDiagnostics struct {
	Reason                 string          `json:"reason,omitempty"`
	BlockedByAssignmentID  string          `json:"blocked_by_assignment_id,omitempty"`
	BlockerAgentID         string          `json:"blocker_agent_id,omitempty"`
	BlockerRuntimeProvider string          `json:"blocker_runtime_provider,omitempty"`
	BlockerState           AssignmentState `json:"blocker_state,omitempty"`
	BlockerUpdatedAt       time.Time       `json:"blocker_updated_at,omitempty"`
}

type AIAgentTaskThreadFailureDiagnostics struct {
	ResultStatus    string `json:"result_status,omitempty"`
	FailureCategory string `json:"failure_category,omitempty"`
	Message         string `json:"message,omitempty"`
}

type AIAgentTaskThreadRecord struct {
	ThreadID           string                               `json:"thread_id"`
	ConversationID     string                               `json:"conversation_id,omitempty"`
	ParentThreadID     string                               `json:"parent_thread_id,omitempty"`
	TaskID             string                               `json:"task_id"`
	AssignmentID       string                               `json:"assignment_id,omitempty"`
	AgentID            string                               `json:"agent_id"`
	AgentSnapshot      *AIAgentTaskThreadAgentSnapshot      `json:"agent_snapshot,omitempty"`
	RunID              string                               `json:"run_id"`
	SourceCommentID    string                               `json:"source_comment_id,omitempty"`
	SourceMessageID    string                               `json:"source_message_id,omitempty"`
	WorkStatus         AgentWorkStatus                      `json:"work_status"`
	AssignmentState    AgentAssignmentState                 `json:"assignment_state"`
	QueueDiagnostics   *AIAgentTaskThreadQueueDiagnostics   `json:"queue_diagnostics,omitempty"`
	FailureDiagnostics *AIAgentTaskThreadFailureDiagnostics `json:"failure_diagnostics,omitempty"`
	CommentKind        AgentTaskCommentKind                 `json:"comment_kind"`
	Message            string                               `json:"message"`
	ResultMessage      string                               `json:"result_message,omitempty"`
	StartedAt          time.Time                            `json:"started_at,omitempty"`
	CompletedAt        time.Time                            `json:"completed_at,omitempty"`
	Lines              []AgentThreadProgressLine            `json:"lines"`
}

type AIAgentTaskThreadCollectionResponse struct {
	SchemaVersion string                       `json:"schema_version"`
	TaskID        string                       `json:"task_id"`
	Threads       []AIAgentTaskThreadRecord    `json:"threads"`
	ActiveStream  *AIAgentTaskThreadStreamLink `json:"active_stream,omitempty"`
}

type AIAgentTaskThreadStreamSubscriptionResponse struct {
	SchemaVersion       string                          `json:"schema_version"`
	TaskID              string                          `json:"task_id"`
	Stream              AIAgentTaskEventStreamLink      `json:"stream"`
	ActiveThreadFilters []AIAgentTaskThreadStreamTarget `json:"active_thread_filters"`
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
	EventType          string                               `json:"event_type"`
	SchemaVersion      string                               `json:"schema_version"`
	AgentID            string                               `json:"agent_id"`
	TaskID             string                               `json:"task_id,omitempty"`
	AssignmentID       string                               `json:"assignment_id,omitempty"`
	ThreadID           string                               `json:"thread_id,omitempty"`
	RunID              string                               `json:"run_id,omitempty"`
	WorkStatus         AgentWorkStatus                      `json:"work_status"`
	AssignmentState    AgentAssignmentState                 `json:"assignment_state,omitempty"`
	CommentKind        AgentTaskCommentKind                 `json:"comment_kind,omitempty"`
	ResultMessage      string                               `json:"result_message,omitempty"`
	FailureDiagnostics *AIAgentTaskThreadFailureDiagnostics `json:"failure_diagnostics,omitempty"`
}

type AgentThreadProgressLine struct {
	Seq         int               `json:"seq"`
	Message     string            `json:"message"`
	MessageCode int               `json:"message_code,omitempty"`
	MessageKey  string            `json:"message_key,omitempty"`
	MessageArgs map[string]string `json:"message_args,omitempty"`
	ObservedAt  time.Time         `json:"observed_at,omitempty"`
}

func (line *AgentThreadProgressLine) UnmarshalJSON(data []byte) error {
	var wire struct {
		Seq         int               `json:"seq"`
		Message     string            `json:"message"`
		MessageCode int               `json:"message_code,omitempty"`
		MessageKey  string            `json:"message_key,omitempty"`
		MessageArgs map[string]string `json:"message_args,omitempty"`
		ObservedAt  time.Time         `json:"observed_at,omitempty"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	line.Seq = wire.Seq
	line.Message = wire.Message
	line.MessageCode = wire.MessageCode
	line.MessageKey = wire.MessageKey
	line.MessageArgs = wire.MessageArgs
	line.ObservedAt = wire.ObservedAt
	return nil
}

func (line AgentThreadProgressLine) MarshalJSON() ([]byte, error) {
	wire := struct {
		Seq        int       `json:"seq"`
		Message    string    `json:"message"`
		ObservedAt time.Time `json:"observed_at,omitempty"`
	}{
		Seq:        line.Seq,
		Message:    line.Message,
		ObservedAt: line.ObservedAt,
	}
	return json.Marshal(wire)
}

type AgentThreadProgressEvent struct {
	EventType       string                    `json:"event_type"`
	SchemaVersion   string                    `json:"schema_version"`
	AgentID         string                    `json:"agent_id"`
	TaskID          string                    `json:"task_id"`
	AssignmentID    string                    `json:"assignment_id,omitempty"`
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
