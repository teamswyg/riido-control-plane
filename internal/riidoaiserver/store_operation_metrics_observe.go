package riidoaiserver

func (m *StoreOperationMetrics) ObserveStoreOperation(obs StoreOperationObservation) {
	if m == nil {
		return
	}
	operation := obs.Operation.String()
	observedAt := metricsObservedAt(obs.ObservedAt)
	bucketStart := metricsBucketStart(observedAt)
	elapsedMS := durationMilliseconds(obs.Duration)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.pruneLocked(observedAt)
	bucket := m.bucketLocked(bucketStart)
	bucket.callsTotal++
	if storeOperationFailed(obs.Err) {
		bucket.errorsTotal++
	}
	bucket.latency.observe(elapsedMS)
	if observedAt.After(bucket.lastObservedAt) {
		bucket.lastObservedAt = observedAt
	}

	state := bucket.byOperation[operation]
	metric := state.metric
	metric.Operation = operation
	metric.CallsTotal++
	if storeOperationFailed(obs.Err) {
		metric.ErrorsTotal++
	}
	metric.LatencySamplesTotal++
	metric.LatencyTotalMilliseconds += elapsedMS
	if elapsedMS > metric.LatencyMaxMilliseconds {
		metric.LatencyMaxMilliseconds = elapsedMS
	}
	metric.LatencyLastMilliseconds = elapsedMS
	state.metric = metric
	if observedAt.After(state.lastObservedAt) {
		state.lastObservedAt = observedAt
	}
	bucket.byOperation[operation] = state
}
