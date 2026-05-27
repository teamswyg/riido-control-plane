package riidoaiserver

import "time"

const MetricsSchemaVersion = "riido-ai-server-metrics.v1"

type AssignRequest struct {
	ComponentID     string `json:"component_id"`
	AgentID         string `json:"agent_id"`
	RuntimeProvider string `json:"runtime_provider"`
	Prompt          string `json:"prompt"`
	CreatedBy       string `json:"created_by,omitempty"`
}

type Assignment struct {
	ID                    string          `json:"assignment_id"`
	TaskID                string          `json:"task_id"`
	ComponentID           string          `json:"component_id"`
	AgentID               string          `json:"agent_id"`
	RuntimeProvider       string          `json:"runtime_provider"`
	Prompt                string          `json:"prompt"`
	State                 AssignmentState `json:"state"`
	LeaseToken            string          `json:"lease_token,omitempty"`
	ReplacesAssignmentID  string          `json:"replaces_assignment_id,omitempty"`
	BlockedByAssignmentID string          `json:"blocked_by_assignment_id,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type PollRequest struct {
	DaemonID  string `json:"daemon_id"`
	DeviceID  string `json:"device_id"`
	RuntimeID string `json:"runtime_id"`
}

type AgentHeartbeatRequest struct {
	DaemonID            string   `json:"daemon_id"`
	DeviceID            string   `json:"device_id"`
	RuntimeID           string   `json:"runtime_id"`
	RunningTaskIDs      []string `json:"running_task_ids,omitempty"`
	ActiveAssignmentIDs []string `json:"active_assignment_ids,omitempty"`
}

type PollResponse struct {
	SchemaVersion string      `json:"schema_version"`
	Action        PollAction  `json:"action"`
	Assignment    *Assignment `json:"assignment,omitempty"`
}

type AgentHeartbeatResponse struct {
	SchemaVersion        string       `json:"schema_version"`
	RefreshedAssignments []Assignment `json:"refreshed_assignments,omitempty"`
}

type AgentEventRequest struct {
	AssignmentID string            `json:"assignment_id"`
	TaskID       string            `json:"task_id"`
	DaemonID     string            `json:"daemon_id,omitempty"`
	DeviceID     string            `json:"device_id,omitempty"`
	RuntimeID    string            `json:"runtime_id,omitempty"`
	State        AssignmentState   `json:"state,omitempty"`
	EventType    string            `json:"event_type,omitempty"`
	Message      string            `json:"message,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type AgentEventResponse struct {
	SchemaVersion string      `json:"schema_version"`
	Assignment    *Assignment `json:"assignment,omitempty"`
	Event         TaskEvent   `json:"event"`
}

type TaskEvent struct {
	Seq          int64             `json:"seq"`
	TaskID       string            `json:"task_id"`
	AssignmentID string            `json:"assignment_id"`
	AgentID      string            `json:"agent_id"`
	Type         string            `json:"type"`
	State        AssignmentState   `json:"state,omitempty"`
	Message      string            `json:"message,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	At           time.Time         `json:"at"`
}

type Health struct {
	SchemaVersion string `json:"schema_version"`
	Status        string `json:"status"`
}

type MetricsSnapshot struct {
	SchemaVersion                       string                  `json:"schema_version"`
	GeneratedAt                         time.Time               `json:"generated_at"`
	TasksTotal                          int                     `json:"tasks_total"`
	AssignmentsTotal                    int                     `json:"assignments_total"`
	AssignmentsByState                  map[AssignmentState]int `json:"assignments_by_state"`
	PollRequestsTotal                   int64                   `json:"poll_requests_total"`
	PollActionsTotal                    map[PollAction]int64    `json:"poll_actions_total"`
	AgentEventsTotal                    int64                   `json:"agent_events_total"`
	TaskEventsTotal                     int64                   `json:"task_events_total"`
	SSESubscribers                      int                     `json:"sse_subscribers"`
	OutboxErrorsTotal                   int64                   `json:"outbox_errors_total"`
	EventAppendLatencySamplesTotal      int64                   `json:"event_append_latency_samples_total"`
	EventAppendLatencyTotalMilliseconds int64                   `json:"event_append_latency_total_ms"`
	EventAppendLatencyMaxMilliseconds   int64                   `json:"event_append_latency_max_ms"`
	EventAppendLatencyLastMilliseconds  int64                   `json:"event_append_latency_last_ms"`
}
