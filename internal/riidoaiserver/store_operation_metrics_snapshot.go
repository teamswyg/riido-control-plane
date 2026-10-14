package riidoaiserver

import "time"

func (m *StoreOperationMetrics) ApplyToMetricsSnapshot(snapshot MetricsSnapshot) MetricsSnapshot {
	if m == nil {
		return snapshot
	}
	callsTotal, errorsTotal, latency, operations := m.snapshot()
	snapshot.StoreOperationCallsTotal = callsTotal
	snapshot.StoreOperationErrorsTotal = errorsTotal
	snapshot.StoreOperationLatencySamplesTotal = latency.samplesTotal
	snapshot.StoreOperationLatencyTotalMilliseconds = latency.totalMilliseconds
	snapshot.StoreOperationLatencyMaxMilliseconds = latency.maxMilliseconds
	snapshot.StoreOperationLatencyLastMilliseconds = latency.lastMilliseconds
	snapshot.StoreOperations = operations
	return snapshot
}

func (m *StoreOperationMetrics) snapshot() (int64, int64, storeOperationLatencyMetrics, []StoreOperationMetric) {
	callsTotal := int64(0)
	errorsTotal := int64(0)
	latency := storeOperationLatencyMetrics{}
	latencyLastObservedAt := time.Time{}
	byOperation := map[string]storeOperationMetricState{}

	m.mu.Lock()
	m.pruneLocked(time.Now())
	for _, bucket := range m.buckets {
		callsTotal += bucket.callsTotal
		errorsTotal += bucket.errorsTotal
		latency.samplesTotal += bucket.latency.samplesTotal
		latency.totalMilliseconds += bucket.latency.totalMilliseconds
		if bucket.latency.maxMilliseconds > latency.maxMilliseconds {
			latency.maxMilliseconds = bucket.latency.maxMilliseconds
		}
		if bucket.lastObservedAt.After(latencyLastObservedAt) {
			latencyLastObservedAt = bucket.lastObservedAt
			latency.lastMilliseconds = bucket.latency.lastMilliseconds
		}
		mergeStoreOperationMetrics(byOperation, *bucket)
	}
	m.mu.Unlock()

	operations := make([]StoreOperationMetric, 0, len(byOperation))
	for _, state := range byOperation {
		operations = append(operations, state.metric)
	}
	sortStoreOperationMetrics(operations)
	if len(operations) > metricsBreakdownLimit {
		operations = operations[:metricsBreakdownLimit]
	}
	return callsTotal, errorsTotal, latency, operations
}
