package riidoaiserver

import (
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

const MetricsSchemaVersion = "riido-ai-server-metrics.v1"

type (
	AssignRequest          = assignmentcontract.AssignRequest
	Assignment             = assignmentcontract.Assignment
	AssignmentWorktree     = assignmentcontract.AssignmentWorktree
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
	SchemaVersion                                     string                  `json:"schema_version"`
	GeneratedAt                                       time.Time               `json:"generated_at"`
	TasksTotal                                        int                     `json:"tasks_total"`
	AssignmentsTotal                                  int                     `json:"assignments_total"`
	AssignmentsByState                                map[AssignmentState]int `json:"assignments_by_state"`
	PollRequestsTotal                                 int64                   `json:"poll_requests_total"`
	PollActionsTotal                                  map[PollAction]int64    `json:"poll_actions_total"`
	AgentEventsTotal                                  int64                   `json:"agent_events_total"`
	TaskEventsTotal                                   int64                   `json:"task_events_total"`
	SSESubscribers                                    int                     `json:"sse_subscribers"`
	SSEStreamsActive                                  int64                   `json:"sse_streams_active,omitempty"`
	SSEStreamsOpenedTotal                             int64                   `json:"sse_streams_opened_total,omitempty"`
	SSEStreamsClosedTotal                             int64                   `json:"sse_streams_closed_total,omitempty"`
	SSEStreamTTFBSamplesTotal                         int64                   `json:"sse_stream_ttfb_samples_total,omitempty"`
	SSEStreamTTFBTotalMilliseconds                    int64                   `json:"sse_stream_ttfb_total_ms,omitempty"`
	SSEStreamTTFBMaxMilliseconds                      int64                   `json:"sse_stream_ttfb_max_ms,omitempty"`
	SSEStreamTTFBLastMilliseconds                     int64                   `json:"sse_stream_ttfb_last_ms,omitempty"`
	SSEStreamDurationSamplesTotal                     int64                   `json:"sse_stream_duration_samples_total,omitempty"`
	SSEStreamDurationTotalMilliseconds                int64                   `json:"sse_stream_duration_total_ms,omitempty"`
	SSEStreamDurationMaxMilliseconds                  int64                   `json:"sse_stream_duration_max_ms,omitempty"`
	SSEStreamDurationLastMilliseconds                 int64                   `json:"sse_stream_duration_last_ms,omitempty"`
	SSEStreams                                        []SSEStreamMetric       `json:"sse_streams,omitempty"`
	OutboxErrorsTotal                                 int64                   `json:"outbox_errors_total"`
	EventAppendLatencySamplesTotal                    int64                   `json:"event_append_latency_samples_total"`
	EventAppendLatencyTotalMilliseconds               int64                   `json:"event_append_latency_total_ms"`
	EventAppendLatencyMaxMilliseconds                 int64                   `json:"event_append_latency_max_ms"`
	EventAppendLatencyLastMilliseconds                int64                   `json:"event_append_latency_last_ms"`
	HTTPRequestsTotal                                 int64                   `json:"http_requests_total,omitempty"`
	HTTPResponsesByStatus                             map[int]int64           `json:"http_responses_by_status,omitempty"`
	HTTPRequestLatencySamplesTotal                    int64                   `json:"http_request_latency_samples_total,omitempty"`
	HTTPRequestLatencyTotalMilliseconds               int64                   `json:"http_request_latency_total_ms,omitempty"`
	HTTPRequestLatencyMaxMilliseconds                 int64                   `json:"http_request_latency_max_ms,omitempty"`
	HTTPRequestLatencyLastMilliseconds                int64                   `json:"http_request_latency_last_ms,omitempty"`
	HTTPRequestsDaemonTotal                           int64                   `json:"http_requests_daemon_total,omitempty"`
	HTTPRequestsClientAppTotal                        int64                   `json:"http_requests_client_app_total,omitempty"`
	HTTPRequestsDesktopTotal                          int64                   `json:"http_requests_desktop_total,omitempty"`
	HTTPRequestsDesktopCandidateTotal                 int64                   `json:"http_requests_desktop_candidate_total,omitempty"`
	HTTPRequestsComponentIntegrationTotal             int64                   `json:"http_requests_component_integration_total,omitempty"`
	HTTPRequestsUnknownSurfaceTotal                   int64                   `json:"http_requests_unknown_surface_total,omitempty"`
	HTTPTransactions                                  []HTTPTransactionMetric `json:"http_transactions,omitempty"`
	StoreOperationCallsTotal                          int64                   `json:"store_operation_calls_total,omitempty"`
	StoreOperationErrorsTotal                         int64                   `json:"store_operation_errors_total,omitempty"`
	StoreOperationLatencySamplesTotal                 int64                   `json:"store_operation_latency_samples_total,omitempty"`
	StoreOperationLatencyTotalMilliseconds            int64                   `json:"store_operation_latency_total_ms,omitempty"`
	StoreOperationLatencyMaxMilliseconds              int64                   `json:"store_operation_latency_max_ms,omitempty"`
	StoreOperationLatencyLastMilliseconds             int64                   `json:"store_operation_latency_last_ms,omitempty"`
	StoreOperations                                   []StoreOperationMetric  `json:"store_operations,omitempty"`
	AIAgentClientSnapshotLoadCallsTotal               int64                   `json:"ai_agent_client_snapshot_load_calls_total,omitempty"`
	AIAgentClientSnapshotLoadErrorsTotal              int64                   `json:"ai_agent_client_snapshot_load_errors_total,omitempty"`
	AIAgentClientSnapshotLoadBytesTotal               int64                   `json:"ai_agent_client_snapshot_load_bytes_total,omitempty"`
	AIAgentClientSnapshotLoadBytesMax                 int64                   `json:"ai_agent_client_snapshot_load_bytes_max,omitempty"`
	AIAgentClientSnapshotLoadBytesLast                int64                   `json:"ai_agent_client_snapshot_load_bytes_last,omitempty"`
	AIAgentClientSnapshotLoadLatencySamplesTotal      int64                   `json:"ai_agent_client_snapshot_load_latency_samples_total,omitempty"`
	AIAgentClientSnapshotLoadLatencyTotalMilliseconds int64                   `json:"ai_agent_client_snapshot_load_latency_total_ms,omitempty"`
	AIAgentClientSnapshotLoadLatencyMaxMilliseconds   int64                   `json:"ai_agent_client_snapshot_load_latency_max_ms,omitempty"`
	AIAgentClientSnapshotLoadLatencyLastMilliseconds  int64                   `json:"ai_agent_client_snapshot_load_latency_last_ms,omitempty"`
	AIAgentClientSnapshotSaveCallsTotal               int64                   `json:"ai_agent_client_snapshot_save_calls_total,omitempty"`
	AIAgentClientSnapshotSaveErrorsTotal              int64                   `json:"ai_agent_client_snapshot_save_errors_total,omitempty"`
	AIAgentClientSnapshotSaveBytesTotal               int64                   `json:"ai_agent_client_snapshot_save_bytes_total,omitempty"`
	AIAgentClientSnapshotSaveBytesMax                 int64                   `json:"ai_agent_client_snapshot_save_bytes_max,omitempty"`
	AIAgentClientSnapshotSaveBytesLast                int64                   `json:"ai_agent_client_snapshot_save_bytes_last,omitempty"`
	AIAgentClientSnapshotSaveLatencySamplesTotal      int64                   `json:"ai_agent_client_snapshot_save_latency_samples_total,omitempty"`
	AIAgentClientSnapshotSaveLatencyTotalMilliseconds int64                   `json:"ai_agent_client_snapshot_save_latency_total_ms,omitempty"`
	AIAgentClientSnapshotSaveLatencyMaxMilliseconds   int64                   `json:"ai_agent_client_snapshot_save_latency_max_ms,omitempty"`
	AIAgentClientSnapshotSaveLatencyLastMilliseconds  int64                   `json:"ai_agent_client_snapshot_save_latency_last_ms,omitempty"`
}
