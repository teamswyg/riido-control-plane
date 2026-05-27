package riidoaiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestWriteCloudWatchEMF(t *testing.T) {
	var buf bytes.Buffer
	snapshot := MetricsSnapshot{
		SchemaVersion:                       MetricsSchemaVersion,
		GeneratedAt:                         time.Unix(123, 456000000).UTC(),
		TasksTotal:                          2,
		AssignmentsTotal:                    3,
		AssignmentsByState:                  map[AssignmentState]int{AssignmentQueued: 1, AssignmentRunning: 2},
		PollRequestsTotal:                   5,
		PollActionsTotal:                    map[PollAction]int64{PollStart: 2, PollNone: 3},
		AgentEventsTotal:                    7,
		TaskEventsTotal:                     11,
		SSESubscribers:                      13,
		OutboxErrorsTotal:                   17,
		EventAppendLatencySamplesTotal:      19,
		EventAppendLatencyTotalMilliseconds: 230,
		EventAppendLatencyMaxMilliseconds:   89,
		EventAppendLatencyLastMilliseconds:  34,
	}
	if err := WriteCloudWatchEMF(&buf, CloudWatchEMFConfig{}, snapshot); err != nil {
		t.Fatalf("WriteCloudWatchEMF: %v", err)
	}
	var out struct {
		AWS struct {
			Timestamp         int64 `json:"Timestamp"`
			CloudWatchMetrics []struct {
				Namespace  string     `json:"Namespace"`
				Dimensions [][]string `json:"Dimensions"`
				Metrics    []struct {
					Name string `json:"Name"`
					Unit string `json:"Unit"`
				} `json:"Metrics"`
			} `json:"CloudWatchMetrics"`
		} `json:"_aws"`
		SchemaVersion                       string `json:"schema_version"`
		Service                             string `json:"service"`
		TasksTotal                          int    `json:"tasks_total"`
		AssignmentsQueued                   int    `json:"assignments_queued"`
		AssignmentsRunning                  int    `json:"assignments_running"`
		PollStartTotal                      int64  `json:"poll_start_total"`
		SSESubscribers                      int    `json:"sse_subscribers"`
		OutboxErrorsTotal                   int64  `json:"outbox_errors_total"`
		EventAppendLatencySamplesTotal      int64  `json:"event_append_latency_samples_total"`
		EventAppendLatencyTotalMilliseconds int64  `json:"event_append_latency_total_ms"`
		EventAppendLatencyMaxMilliseconds   int64  `json:"event_append_latency_max_ms"`
		EventAppendLatencyLastMilliseconds  int64  `json:"event_append_latency_last_ms"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("decode emf: %v\n%s", err, buf.String())
	}
	if out.AWS.Timestamp != 123456 || out.AWS.CloudWatchMetrics[0].Namespace != defaultCloudWatchNamespace {
		t.Fatalf("emf metadata = %+v", out.AWS)
	}
	if out.AWS.CloudWatchMetrics[0].Dimensions[0][0] != "service" || len(out.AWS.CloudWatchMetrics[0].Metrics) == 0 {
		t.Fatalf("emf metric specs = %+v", out.AWS.CloudWatchMetrics[0])
	}
	if out.SchemaVersion != MetricsSchemaVersion || out.Service != defaultCloudWatchServiceName {
		t.Fatalf("emf identity = %+v", out)
	}
	if out.TasksTotal != 2 || out.AssignmentsQueued != 1 || out.AssignmentsRunning != 2 || out.PollStartTotal != 2 || out.SSESubscribers != 13 || out.OutboxErrorsTotal != 17 {
		t.Fatalf("emf counters = %+v", out)
	}
	if out.EventAppendLatencySamplesTotal != 19 || out.EventAppendLatencyTotalMilliseconds != 230 || out.EventAppendLatencyMaxMilliseconds != 89 || out.EventAppendLatencyLastMilliseconds != 34 {
		t.Fatalf("emf latency counters = %+v", out)
	}
	metricUnits := map[string]string{}
	for _, spec := range out.AWS.CloudWatchMetrics[0].Metrics {
		metricUnits[spec.Name] = spec.Unit
	}
	if metricUnits["event_append_latency_samples_total"] != "Count" || metricUnits["event_append_latency_max_ms"] != "Milliseconds" {
		t.Fatalf("emf latency metric units = %+v", metricUnits)
	}
}

func TestPublishCloudWatchEMFUsesStoreSnapshot(t *testing.T) {
	now := time.Unix(1000, 0).UTC()
	store := NewStoreWithClock(func() time.Time { return now })
	defer store.Close()
	ctx := context.Background()
	if _, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-1",
		AgentID:         "jykim1",
		RuntimeProvider: "codex",
		Prompt:          "hello",
	}); err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	var buf bytes.Buffer
	if err := PublishCloudWatchEMF(ctx, store, CloudWatchEMFConfig{Writer: &buf, Namespace: "Riido/Test", ServiceName: "test-service"}); err != nil {
		t.Fatalf("PublishCloudWatchEMF: %v", err)
	}
	var out struct {
		AWS struct {
			CloudWatchMetrics []struct {
				Namespace string `json:"Namespace"`
			} `json:"CloudWatchMetrics"`
		} `json:"_aws"`
		Service          string `json:"service"`
		AssignmentsTotal int    `json:"assignments_total"`
		TaskEventsTotal  int64  `json:"task_events_total"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("decode emf: %v", err)
	}
	if out.AWS.CloudWatchMetrics[0].Namespace != "Riido/Test" || out.Service != "test-service" {
		t.Fatalf("emf identity = %+v", out)
	}
	if out.AssignmentsTotal != 1 || out.TaskEventsTotal != 1 {
		t.Fatalf("store counters = %+v", out)
	}
}
