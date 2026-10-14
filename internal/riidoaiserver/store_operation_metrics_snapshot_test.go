package riidoaiserver

import (
	"errors"
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
