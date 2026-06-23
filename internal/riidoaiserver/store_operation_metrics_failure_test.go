package riidoaiserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

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
	poll := storeOperationMetricsByName(snapshot.StoreOperations)[StoreOperationPollAssignment.String()]
	if poll.CallsTotal != 1 || poll.ErrorsTotal != 0 || poll.LatencySamplesTotal != 1 {
		t.Fatalf("poll metric should keep call/latency and suppress binding validation error: %+v", poll)
	}
}
