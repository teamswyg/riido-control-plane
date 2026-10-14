package riidoaiserver

import (
	"fmt"
	"testing"
	"time"
)

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
