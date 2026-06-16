package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

const (
	defaultCloudWatchNamespace   = "Riido/RiidoAIServer"
	defaultCloudWatchServiceName = "riido_ai_server"
)

type CloudWatchEMFConfig struct {
	Namespace   string
	ServiceName string
	Writer      io.Writer
}

type cloudWatchEMFEnvelope struct {
	AWS                                               cloudWatchEMFMetadata   `json:"_aws"`
	SchemaVersion                                     string                  `json:"schema_version"`
	Service                                           string                  `json:"service"`
	TasksTotal                                        int                     `json:"tasks_total"`
	AssignmentsTotal                                  int                     `json:"assignments_total"`
	AssignmentsQueued                                 int                     `json:"assignments_queued"`
	AssignmentsLeased                                 int                     `json:"assignments_leased"`
	AssignmentsReady                                  int                     `json:"assignments_ready"`
	AssignmentsRunning                                int                     `json:"assignments_running"`
	AssignmentsCancelling                             int                     `json:"assignments_cancelling"`
	AssignmentsCancelled                              int                     `json:"assignments_cancelled"`
	AssignmentsCompleted                              int                     `json:"assignments_completed"`
	AssignmentsFailed                                 int                     `json:"assignments_failed"`
	PollRequestsTotal                                 int64                   `json:"poll_requests_total"`
	PollNoneTotal                                     int64                   `json:"poll_none_total"`
	PollStartTotal                                    int64                   `json:"poll_start_total"`
	PollCancelTotal                                   int64                   `json:"poll_cancel_total"`
	PollActiveTotal                                   int64                   `json:"poll_active_total"`
	AgentEventsTotal                                  int64                   `json:"agent_events_total"`
	TaskEventsTotal                                   int64                   `json:"task_events_total"`
	SSESubscribers                                    int                     `json:"sse_subscribers"`
	OutboxErrorsTotal                                 int64                   `json:"outbox_errors_total"`
	EventAppendLatencySamplesTotal                    int64                   `json:"event_append_latency_samples_total"`
	EventAppendLatencyTotalMilliseconds               int64                   `json:"event_append_latency_total_ms"`
	EventAppendLatencyMaxMilliseconds                 int64                   `json:"event_append_latency_max_ms"`
	EventAppendLatencyLastMilliseconds                int64                   `json:"event_append_latency_last_ms"`
	HTTPRequestsTotal                                 int64                   `json:"http_requests_total"`
	HTTPResponse2xxTotal                              int64                   `json:"http_response_2xx_total"`
	HTTPResponse3xxTotal                              int64                   `json:"http_response_3xx_total"`
	HTTPResponse4xxTotal                              int64                   `json:"http_response_4xx_total"`
	HTTPResponse5xxTotal                              int64                   `json:"http_response_5xx_total"`
	HTTPRequestLatencySamplesTotal                    int64                   `json:"http_request_latency_samples_total"`
	HTTPRequestLatencyTotalMilliseconds               int64                   `json:"http_request_latency_total_ms"`
	HTTPRequestLatencyMaxMilliseconds                 int64                   `json:"http_request_latency_max_ms"`
	HTTPRequestLatencyLastMilliseconds                int64                   `json:"http_request_latency_last_ms"`
	HTTPTransactions                                  []HTTPTransactionMetric `json:"http_transactions,omitempty"`
	StoreOperationCallsTotal                          int64                   `json:"store_operation_calls_total"`
	StoreOperationErrorsTotal                         int64                   `json:"store_operation_errors_total"`
	StoreOperationLatencySamplesTotal                 int64                   `json:"store_operation_latency_samples_total"`
	StoreOperationLatencyTotalMilliseconds            int64                   `json:"store_operation_latency_total_ms"`
	StoreOperationLatencyMaxMilliseconds              int64                   `json:"store_operation_latency_max_ms"`
	StoreOperationLatencyLastMilliseconds             int64                   `json:"store_operation_latency_last_ms"`
	StoreOperations                                   []StoreOperationMetric  `json:"store_operations,omitempty"`
	AIAgentClientSnapshotLoadCallsTotal               int64                   `json:"ai_agent_client_snapshot_load_calls_total"`
	AIAgentClientSnapshotLoadErrorsTotal              int64                   `json:"ai_agent_client_snapshot_load_errors_total"`
	AIAgentClientSnapshotLoadBytesTotal               int64                   `json:"ai_agent_client_snapshot_load_bytes_total"`
	AIAgentClientSnapshotLoadBytesMax                 int64                   `json:"ai_agent_client_snapshot_load_bytes_max"`
	AIAgentClientSnapshotLoadBytesLast                int64                   `json:"ai_agent_client_snapshot_load_bytes_last"`
	AIAgentClientSnapshotLoadLatencySamplesTotal      int64                   `json:"ai_agent_client_snapshot_load_latency_samples_total"`
	AIAgentClientSnapshotLoadLatencyTotalMilliseconds int64                   `json:"ai_agent_client_snapshot_load_latency_total_ms"`
	AIAgentClientSnapshotLoadLatencyMaxMilliseconds   int64                   `json:"ai_agent_client_snapshot_load_latency_max_ms"`
	AIAgentClientSnapshotLoadLatencyLastMilliseconds  int64                   `json:"ai_agent_client_snapshot_load_latency_last_ms"`
	AIAgentClientSnapshotSaveCallsTotal               int64                   `json:"ai_agent_client_snapshot_save_calls_total"`
	AIAgentClientSnapshotSaveErrorsTotal              int64                   `json:"ai_agent_client_snapshot_save_errors_total"`
	AIAgentClientSnapshotSaveBytesTotal               int64                   `json:"ai_agent_client_snapshot_save_bytes_total"`
	AIAgentClientSnapshotSaveBytesMax                 int64                   `json:"ai_agent_client_snapshot_save_bytes_max"`
	AIAgentClientSnapshotSaveBytesLast                int64                   `json:"ai_agent_client_snapshot_save_bytes_last"`
	AIAgentClientSnapshotSaveLatencySamplesTotal      int64                   `json:"ai_agent_client_snapshot_save_latency_samples_total"`
	AIAgentClientSnapshotSaveLatencyTotalMilliseconds int64                   `json:"ai_agent_client_snapshot_save_latency_total_ms"`
	AIAgentClientSnapshotSaveLatencyMaxMilliseconds   int64                   `json:"ai_agent_client_snapshot_save_latency_max_ms"`
	AIAgentClientSnapshotSaveLatencyLastMilliseconds  int64                   `json:"ai_agent_client_snapshot_save_latency_last_ms"`
}

type cloudWatchEMFMetadata struct {
	Timestamp         int64                   `json:"Timestamp"`
	CloudWatchMetrics []cloudWatchMetricGroup `json:"CloudWatchMetrics"`
}

type cloudWatchMetricGroup struct {
	Namespace  string                 `json:"Namespace"`
	Dimensions [][]string             `json:"Dimensions"`
	Metrics    []cloudWatchMetricSpec `json:"Metrics"`
}

type cloudWatchMetricSpec struct {
	Name string `json:"Name"`
	Unit string `json:"Unit"`
}

func PublishCloudWatchEMF(ctx context.Context, metrics MetricsReader, config CloudWatchEMFConfig) error {
	if metrics == nil {
		return errors.New("riidoaiserver: metrics reader is required")
	}
	snapshot, err := metrics.Metrics(ctx)
	if err != nil {
		return err
	}
	return WriteCloudWatchEMF(config.Writer, normalizeCloudWatchEMFConfig(config), snapshot)
}

func RunCloudWatchEMFPublisher(ctx context.Context, metrics MetricsReader, interval time.Duration, config CloudWatchEMFConfig) error {
	if interval <= 0 {
		return errors.New("riidoaiserver: metrics interval must be positive")
	}
	if err := PublishCloudWatchEMF(ctx, metrics, config); err != nil {
		return err
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := PublishCloudWatchEMF(ctx, metrics, config); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func WriteCloudWatchEMF(w io.Writer, config CloudWatchEMFConfig, snapshot MetricsSnapshot) error {
	if w == nil {
		return errors.New("riidoaiserver: metrics writer is required")
	}
	config = normalizeCloudWatchEMFConfig(config)
	envelope := cloudWatchEMFEnvelope{
		AWS: cloudWatchEMFMetadata{
			Timestamp: snapshot.GeneratedAt.UnixMilli(),
			CloudWatchMetrics: []cloudWatchMetricGroup{{
				Namespace:  config.Namespace,
				Dimensions: [][]string{{"service"}},
				Metrics:    cloudWatchMetricSpecs(),
			}},
		},
		SchemaVersion:                                     snapshot.SchemaVersion,
		Service:                                           config.ServiceName,
		TasksTotal:                                        snapshot.TasksTotal,
		AssignmentsTotal:                                  snapshot.AssignmentsTotal,
		AssignmentsQueued:                                 snapshot.AssignmentsByState[AssignmentQueued],
		AssignmentsLeased:                                 snapshot.AssignmentsByState[AssignmentLeased],
		AssignmentsReady:                                  snapshot.AssignmentsByState[AssignmentReady],
		AssignmentsRunning:                                snapshot.AssignmentsByState[AssignmentRunning],
		AssignmentsCancelling:                             snapshot.AssignmentsByState[AssignmentCancelling],
		AssignmentsCancelled:                              snapshot.AssignmentsByState[AssignmentCancelled],
		AssignmentsCompleted:                              snapshot.AssignmentsByState[AssignmentCompleted],
		AssignmentsFailed:                                 snapshot.AssignmentsByState[AssignmentFailed],
		PollRequestsTotal:                                 snapshot.PollRequestsTotal,
		PollNoneTotal:                                     snapshot.PollActionsTotal[PollNone],
		PollStartTotal:                                    snapshot.PollActionsTotal[PollStart],
		PollCancelTotal:                                   snapshot.PollActionsTotal[PollCancel],
		PollActiveTotal:                                   snapshot.PollActionsTotal[PollActive],
		AgentEventsTotal:                                  snapshot.AgentEventsTotal,
		TaskEventsTotal:                                   snapshot.TaskEventsTotal,
		SSESubscribers:                                    snapshot.SSESubscribers,
		OutboxErrorsTotal:                                 snapshot.OutboxErrorsTotal,
		EventAppendLatencySamplesTotal:                    snapshot.EventAppendLatencySamplesTotal,
		EventAppendLatencyTotalMilliseconds:               snapshot.EventAppendLatencyTotalMilliseconds,
		EventAppendLatencyMaxMilliseconds:                 snapshot.EventAppendLatencyMaxMilliseconds,
		EventAppendLatencyLastMilliseconds:                snapshot.EventAppendLatencyLastMilliseconds,
		HTTPRequestsTotal:                                 snapshot.HTTPRequestsTotal,
		HTTPResponse2xxTotal:                              countHTTPStatusClass(snapshot.HTTPResponsesByStatus, 2),
		HTTPResponse3xxTotal:                              countHTTPStatusClass(snapshot.HTTPResponsesByStatus, 3),
		HTTPResponse4xxTotal:                              countHTTPStatusClass(snapshot.HTTPResponsesByStatus, 4),
		HTTPResponse5xxTotal:                              countHTTPStatusClass(snapshot.HTTPResponsesByStatus, 5),
		HTTPRequestLatencySamplesTotal:                    snapshot.HTTPRequestLatencySamplesTotal,
		HTTPRequestLatencyTotalMilliseconds:               snapshot.HTTPRequestLatencyTotalMilliseconds,
		HTTPRequestLatencyMaxMilliseconds:                 snapshot.HTTPRequestLatencyMaxMilliseconds,
		HTTPRequestLatencyLastMilliseconds:                snapshot.HTTPRequestLatencyLastMilliseconds,
		HTTPTransactions:                                  snapshot.HTTPTransactions,
		StoreOperationCallsTotal:                          snapshot.StoreOperationCallsTotal,
		StoreOperationErrorsTotal:                         snapshot.StoreOperationErrorsTotal,
		StoreOperationLatencySamplesTotal:                 snapshot.StoreOperationLatencySamplesTotal,
		StoreOperationLatencyTotalMilliseconds:            snapshot.StoreOperationLatencyTotalMilliseconds,
		StoreOperationLatencyMaxMilliseconds:              snapshot.StoreOperationLatencyMaxMilliseconds,
		StoreOperationLatencyLastMilliseconds:             snapshot.StoreOperationLatencyLastMilliseconds,
		StoreOperations:                                   snapshot.StoreOperations,
		AIAgentClientSnapshotLoadCallsTotal:               snapshot.AIAgentClientSnapshotLoadCallsTotal,
		AIAgentClientSnapshotLoadErrorsTotal:              snapshot.AIAgentClientSnapshotLoadErrorsTotal,
		AIAgentClientSnapshotLoadBytesTotal:               snapshot.AIAgentClientSnapshotLoadBytesTotal,
		AIAgentClientSnapshotLoadBytesMax:                 snapshot.AIAgentClientSnapshotLoadBytesMax,
		AIAgentClientSnapshotLoadBytesLast:                snapshot.AIAgentClientSnapshotLoadBytesLast,
		AIAgentClientSnapshotLoadLatencySamplesTotal:      snapshot.AIAgentClientSnapshotLoadLatencySamplesTotal,
		AIAgentClientSnapshotLoadLatencyTotalMilliseconds: snapshot.AIAgentClientSnapshotLoadLatencyTotalMilliseconds,
		AIAgentClientSnapshotLoadLatencyMaxMilliseconds:   snapshot.AIAgentClientSnapshotLoadLatencyMaxMilliseconds,
		AIAgentClientSnapshotLoadLatencyLastMilliseconds:  snapshot.AIAgentClientSnapshotLoadLatencyLastMilliseconds,
		AIAgentClientSnapshotSaveCallsTotal:               snapshot.AIAgentClientSnapshotSaveCallsTotal,
		AIAgentClientSnapshotSaveErrorsTotal:              snapshot.AIAgentClientSnapshotSaveErrorsTotal,
		AIAgentClientSnapshotSaveBytesTotal:               snapshot.AIAgentClientSnapshotSaveBytesTotal,
		AIAgentClientSnapshotSaveBytesMax:                 snapshot.AIAgentClientSnapshotSaveBytesMax,
		AIAgentClientSnapshotSaveBytesLast:                snapshot.AIAgentClientSnapshotSaveBytesLast,
		AIAgentClientSnapshotSaveLatencySamplesTotal:      snapshot.AIAgentClientSnapshotSaveLatencySamplesTotal,
		AIAgentClientSnapshotSaveLatencyTotalMilliseconds: snapshot.AIAgentClientSnapshotSaveLatencyTotalMilliseconds,
		AIAgentClientSnapshotSaveLatencyMaxMilliseconds:   snapshot.AIAgentClientSnapshotSaveLatencyMaxMilliseconds,
		AIAgentClientSnapshotSaveLatencyLastMilliseconds:  snapshot.AIAgentClientSnapshotSaveLatencyLastMilliseconds,
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(body))
	return err
}

func normalizeCloudWatchEMFConfig(config CloudWatchEMFConfig) CloudWatchEMFConfig {
	if config.Namespace == "" {
		config.Namespace = defaultCloudWatchNamespace
	}
	if config.ServiceName == "" {
		config.ServiceName = defaultCloudWatchServiceName
	}
	return config
}

func cloudWatchMetricSpecs() []cloudWatchMetricSpec {
	specs := []cloudWatchMetricSpec{
		{Name: "tasks_total", Unit: "Count"},
		{Name: "assignments_total", Unit: "Count"},
		{Name: "assignments_queued", Unit: "Count"},
		{Name: "assignments_leased", Unit: "Count"},
		{Name: "assignments_ready", Unit: "Count"},
		{Name: "assignments_running", Unit: "Count"},
		{Name: "assignments_cancelling", Unit: "Count"},
		{Name: "assignments_cancelled", Unit: "Count"},
		{Name: "assignments_completed", Unit: "Count"},
		{Name: "assignments_failed", Unit: "Count"},
		{Name: "poll_requests_total", Unit: "Count"},
		{Name: "poll_none_total", Unit: "Count"},
		{Name: "poll_start_total", Unit: "Count"},
		{Name: "poll_cancel_total", Unit: "Count"},
		{Name: "poll_active_total", Unit: "Count"},
		{Name: "agent_events_total", Unit: "Count"},
		{Name: "task_events_total", Unit: "Count"},
		{Name: "sse_subscribers", Unit: "Count"},
		{Name: "outbox_errors_total", Unit: "Count"},
		{Name: "event_append_latency_samples_total", Unit: "Count"},
		{Name: "event_append_latency_total_ms", Unit: "Milliseconds"},
		{Name: "event_append_latency_max_ms", Unit: "Milliseconds"},
		{Name: "event_append_latency_last_ms", Unit: "Milliseconds"},
		{Name: "http_requests_total", Unit: "Count"},
		{Name: "http_response_2xx_total", Unit: "Count"},
		{Name: "http_response_3xx_total", Unit: "Count"},
		{Name: "http_response_4xx_total", Unit: "Count"},
		{Name: "http_response_5xx_total", Unit: "Count"},
		{Name: "http_request_latency_samples_total", Unit: "Count"},
		{Name: "http_request_latency_total_ms", Unit: "Milliseconds"},
		{Name: "http_request_latency_max_ms", Unit: "Milliseconds"},
		{Name: "http_request_latency_last_ms", Unit: "Milliseconds"},
		{Name: "store_operation_calls_total", Unit: "Count"},
		{Name: "store_operation_errors_total", Unit: "Count"},
		{Name: "store_operation_latency_samples_total", Unit: "Count"},
		{Name: "store_operation_latency_total_ms", Unit: "Milliseconds"},
		{Name: "store_operation_latency_max_ms", Unit: "Milliseconds"},
		{Name: "store_operation_latency_last_ms", Unit: "Milliseconds"},
		{Name: "ai_agent_client_snapshot_load_calls_total", Unit: "Count"},
		{Name: "ai_agent_client_snapshot_load_errors_total", Unit: "Count"},
		{Name: "ai_agent_client_snapshot_load_bytes_total", Unit: "Bytes"},
		{Name: "ai_agent_client_snapshot_load_bytes_max", Unit: "Bytes"},
		{Name: "ai_agent_client_snapshot_load_bytes_last", Unit: "Bytes"},
		{Name: "ai_agent_client_snapshot_load_latency_samples_total", Unit: "Count"},
		{Name: "ai_agent_client_snapshot_load_latency_total_ms", Unit: "Milliseconds"},
		{Name: "ai_agent_client_snapshot_load_latency_max_ms", Unit: "Milliseconds"},
		{Name: "ai_agent_client_snapshot_load_latency_last_ms", Unit: "Milliseconds"},
		{Name: "ai_agent_client_snapshot_save_calls_total", Unit: "Count"},
		{Name: "ai_agent_client_snapshot_save_errors_total", Unit: "Count"},
		{Name: "ai_agent_client_snapshot_save_bytes_total", Unit: "Bytes"},
		{Name: "ai_agent_client_snapshot_save_bytes_max", Unit: "Bytes"},
		{Name: "ai_agent_client_snapshot_save_bytes_last", Unit: "Bytes"},
		{Name: "ai_agent_client_snapshot_save_latency_samples_total", Unit: "Count"},
		{Name: "ai_agent_client_snapshot_save_latency_total_ms", Unit: "Milliseconds"},
		{Name: "ai_agent_client_snapshot_save_latency_max_ms", Unit: "Milliseconds"},
		{Name: "ai_agent_client_snapshot_save_latency_last_ms", Unit: "Milliseconds"},
	}
	out := make([]cloudWatchMetricSpec, 0, len(specs))
	out = append(out, specs...)
	return out
}

func countHTTPStatusClass(statuses map[int]int64, class int) int64 {
	var total int64
	for statusCode, count := range statuses {
		if statusCode/100 == class {
			total += count
		}
	}
	return total
}
