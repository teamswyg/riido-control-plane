package riidoaiserver

import (
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

const MetricsSchemaVersion = "riido-ai-server-metrics.v1"

type (
	AssignRequest          = assignmentcontract.AssignRequest
	Assignment             = assignmentcontract.Assignment
	PollRequest            = assignmentcontract.PollRequest
	AgentHeartbeatRequest  = assignmentcontract.AgentHeartbeatRequest
	PollResponse           = assignmentcontract.PollResponse
	AgentHeartbeatResponse = assignmentcontract.AgentHeartbeatResponse
	AgentEventRequest      = assignmentcontract.AgentEventRequest
	AgentEventResponse     = assignmentcontract.AgentEventResponse
	TaskEvent              = assignmentcontract.TaskEvent
)

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
