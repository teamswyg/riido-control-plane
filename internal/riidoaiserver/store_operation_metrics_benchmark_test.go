package riidoaiserver

import (
	"testing"
	"time"
)

func BenchmarkStoreOperationMetricsObserve(b *testing.B) {
	metrics := NewStoreOperationMetrics()
	operations := []StoreOperationName{
		StoreOperationPollAssignment,
		StoreOperationLeaseAssignment,
		StoreOperationAppendEvent,
	}
	start := time.Unix(1782250000, 0)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		metrics.ObserveStoreOperation(StoreOperationObservation{
			Operation:  operations[i%len(operations)],
			Duration:   time.Duration(i%17) * time.Millisecond,
			ObservedAt: start.Add(time.Duration(i%120) * time.Second),
		})
	}
}

func BenchmarkStoreOperationMetricsSnapshot(b *testing.B) {
	metrics := NewStoreOperationMetrics()
	seedStoreOperationMetrics(metrics)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = metrics.ApplyToMetricsSnapshot(MetricsSnapshot{})
	}
}

func seedStoreOperationMetrics(metrics *StoreOperationMetrics) {
	operations := []StoreOperationName{
		StoreOperationCreateTask,
		StoreOperationCreateAssignment,
		StoreOperationPollAssignment,
		StoreOperationWaitAssignment,
		StoreOperationLeaseAssignment,
		StoreOperationAppendEvent,
	}
	start := time.Now().Add(-time.Minute)
	for i := 0; i < 512; i++ {
		metrics.ObserveStoreOperation(StoreOperationObservation{
			Operation:  operations[i%len(operations)],
			Duration:   time.Duration(i%23) * time.Millisecond,
			ObservedAt: start.Add(time.Duration(i%120) * time.Second),
		})
	}
}
