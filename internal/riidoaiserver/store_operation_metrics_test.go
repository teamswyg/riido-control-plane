package riidoaiserver

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestStoreOperationMetricsApplyToSnapshot(t *testing.T) {
	metrics := NewStoreOperationMetrics()
	metrics.ObserveStoreOperation(StoreOperationObservation{
		Operation: StoreOperationPollAssignment,
		Duration:  12 * time.Millisecond,
	})
	metrics.ObserveStoreOperation(StoreOperationObservation{
		Operation: StoreOperationPollAssignment,
		Duration:  8 * time.Millisecond,
		Err:       errors.New("boom"),
	})
	metrics.ObserveStoreOperation(StoreOperationObservation{
		Operation: StoreOperationAppendEvent,
		Duration:  3 * time.Millisecond,
	})

	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if snapshot.StoreOperationCallsTotal != 3 || snapshot.StoreOperationErrorsTotal != 1 {
		t.Fatalf("store operation totals = %+v", snapshot)
	}
	if snapshot.StoreOperationLatencySamplesTotal != 3 ||
		snapshot.StoreOperationLatencyTotalMilliseconds != 23 ||
		snapshot.StoreOperationLatencyMaxMilliseconds != 12 ||
		snapshot.StoreOperationLatencyLastMilliseconds != 3 {
		t.Fatalf("store operation latency = %+v", snapshot)
	}
	if len(snapshot.StoreOperations) != 2 {
		t.Fatalf("store operations = %+v", snapshot.StoreOperations)
	}
	if snapshot.StoreOperations[0].Operation != StoreOperationPollAssignment.String() ||
		snapshot.StoreOperations[1].Operation != StoreOperationAppendEvent.String() {
		t.Fatalf("store operations should be sorted by call volume first: %+v", snapshot.StoreOperations)
	}
	poll := snapshot.StoreOperations[0]
	if poll.CallsTotal != 2 || poll.ErrorsTotal != 1 || poll.LatencyMaxMilliseconds != 12 {
		t.Fatalf("poll metric = %+v", poll)
	}
}

func TestStoreOperationMetricsTreatCallerCancellationAsNonStoreFailure(t *testing.T) {
	metrics := NewStoreOperationMetrics()
	metrics.ObserveStoreOperation(StoreOperationObservation{
		Operation: StoreOperationPollAssignment,
		Duration:  time.Millisecond,
		Err:       context.Canceled,
	})

	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if snapshot.StoreOperationCallsTotal != 1 || snapshot.StoreOperationErrorsTotal != 0 {
		t.Fatalf("store operation aggregate should count canceled poll as non-failure: %+v", snapshot)
	}
	if len(snapshot.StoreOperations) != 1 {
		t.Fatalf("store operation breakdown = %+v", snapshot.StoreOperations)
	}
	poll := snapshot.StoreOperations[0]
	if poll.CallsTotal != 1 || poll.ErrorsTotal != 0 || poll.LatencySamplesTotal != 1 {
		t.Fatalf("poll metric should keep call/latency and suppress cancellation error: %+v", poll)
	}
}

func TestStoreMetricsTreatAgentBindingValidationAsNonStoreFailure(t *testing.T) {
	ctx := context.Background()
	registry, err := NewStaticAgentRegistry([]AgentRuntimeBinding{{
		AgentID:         "agent-1",
		DaemonID:        "daemon-1",
		RuntimeID:       "runtime-1",
		RuntimeProvider: "codex",
	}})
	if err != nil {
		t.Fatalf("NewStaticAgentRegistry: %v", err)
	}
	store := NewStoreWithConfig(StoreConfig{AgentRegistry: registry})
	defer store.Close()

	_, err = store.PollAgent(ctx, "agent-1", PollRequest{DaemonID: "daemon-other", RuntimeID: "runtime-1"})
	if err == nil {
		t.Fatal("PollAgent should reject daemon binding mismatch")
	}
	if !errors.Is(err, ErrAgentBindingValidation) {
		t.Fatalf("PollAgent err=%v, want ErrAgentBindingValidation", err)
	}

	snapshot, err := store.Metrics(ctx)
	if err != nil {
		t.Fatalf("Metrics: %v", err)
	}
	if snapshot.StoreOperationCallsTotal != 1 || snapshot.StoreOperationErrorsTotal != 0 {
		t.Fatalf("store operation aggregate should count binding rejection as non-store failure: %+v", snapshot)
	}
	operations := storeOperationMetricsByName(snapshot.StoreOperations)
	poll := operations[StoreOperationPollAssignment.String()]
	if poll.CallsTotal != 1 || poll.ErrorsTotal != 0 || poll.LatencySamplesTotal != 1 {
		t.Fatalf("poll metric should keep call/latency and suppress binding validation error: %+v", poll)
	}
}

func TestStoreOperationMetricsUseRollingWindow(t *testing.T) {
	metrics := NewStoreOperationMetrics()
	old := time.Now().Add(-10 * time.Minute)
	metrics.ObserveStoreOperation(StoreOperationObservation{
		Operation:  StoreOperationAppendEvent,
		Duration:   time.Millisecond,
		ObservedAt: old,
	})
	metrics.ObserveStoreOperation(StoreOperationObservation{
		Operation: StoreOperationPollAssignment,
		Duration:  time.Millisecond,
	})

	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if snapshot.StoreOperationCallsTotal != 1 {
		t.Fatalf("rolling store calls = %d, want 1", snapshot.StoreOperationCallsTotal)
	}
	if len(snapshot.StoreOperations) != 1 || snapshot.StoreOperations[0].Operation != StoreOperationPollAssignment.String() {
		t.Fatalf("rolling store operations = %+v", snapshot.StoreOperations)
	}
}

func TestStoreOperationMetricsLimitBreakdown(t *testing.T) {
	metrics := NewStoreOperationMetrics()
	for i := range metricsBreakdownLimit + 5 {
		metrics.ObserveStoreOperation(StoreOperationObservation{
			Operation: StoreOperationName(fmt.Sprintf("store_generated_%02d", i)),
			Duration:  time.Millisecond,
		})
	}

	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if len(snapshot.StoreOperations) != metricsBreakdownLimit {
		t.Fatalf("store operation breakdown size = %d, want %d", len(snapshot.StoreOperations), metricsBreakdownLimit)
	}
}

func TestStoreMetricsCaptureAssignmentOperationSurface(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	defer store.Close()

	assignment, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-1",
		RuntimeProvider: "codex",
		Prompt:          "ship it",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	poll, err := store.PollAgent(ctx, "agent-1", daemonPollRequest())
	if err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if poll.Action != PollStart || poll.Assignment == nil {
		t.Fatalf("poll = %+v", poll)
	}
	if _, err := store.RecordAgentEvent(ctx, "agent-1", AgentEventRequest{
		AssignmentID: assignment.ID,
		DaemonID:     "daemon-1",
		RuntimeID:    "runtime-1",
		State:        AssignmentRunning,
		EventType:    EventAssignmentRunning,
		Message:      "running",
	}); err != nil {
		t.Fatalf("RecordAgentEvent: %v", err)
	}

	snapshot, err := store.Metrics(ctx)
	if err != nil {
		t.Fatalf("Metrics: %v", err)
	}
	operations := storeOperationMetricsByName(snapshot.StoreOperations)
	for _, operation := range []StoreOperationName{
		StoreOperationCreateTask,
		StoreOperationCreateAssignment,
		StoreOperationPollAssignment,
		StoreOperationLeaseAssignment,
		StoreOperationAppendEvent,
	} {
		metric, ok := operations[operation.String()]
		if !ok {
			t.Fatalf("missing store operation %s in %+v", operation, snapshot.StoreOperations)
		}
		if metric.CallsTotal != 1 || metric.ErrorsTotal != 0 || metric.LatencySamplesTotal != 1 {
			t.Fatalf("store operation %s metric = %+v", operation, metric)
		}
	}
	if snapshot.StoreOperationCallsTotal != 5 || snapshot.StoreOperationLatencySamplesTotal != 5 {
		t.Fatalf("store operation aggregate = %+v", snapshot)
	}
}

func storeOperationMetricsByName(metrics []StoreOperationMetric) map[string]StoreOperationMetric {
	out := make(map[string]StoreOperationMetric, len(metrics))
	for _, metric := range metrics {
		out[metric.Operation] = metric
	}
	return out
}
