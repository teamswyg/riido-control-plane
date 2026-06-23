package riidoaiserver

func mergeStoreOperationMetrics(
	byOperation map[string]storeOperationMetricState,
	bucket storeOperationBucket,
) {
	for operation, state := range bucket.byOperation {
		current := byOperation[operation]
		current.metric.Operation = state.metric.Operation
		current.metric.CallsTotal += state.metric.CallsTotal
		current.metric.ErrorsTotal += state.metric.ErrorsTotal
		current.metric.LatencySamplesTotal += state.metric.LatencySamplesTotal
		current.metric.LatencyTotalMilliseconds += state.metric.LatencyTotalMilliseconds
		if state.metric.LatencyMaxMilliseconds > current.metric.LatencyMaxMilliseconds {
			current.metric.LatencyMaxMilliseconds = state.metric.LatencyMaxMilliseconds
		}
		if state.lastObservedAt.After(current.lastObservedAt) {
			current.lastObservedAt = state.lastObservedAt
			current.metric.LatencyLastMilliseconds = state.metric.LatencyLastMilliseconds
		}
		byOperation[operation] = current
	}
}
