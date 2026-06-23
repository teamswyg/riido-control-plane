package riidoaiserver

func mergeHTTPTransactionBucketState(
	byKey map[httpTransactionKey]httpTransactionMetricState,
	source map[httpTransactionKey]httpTransactionMetricState,
) {
	for key, state := range source {
		current := byKey[key]
		current.metric.Method = state.metric.Method
		current.metric.Route = state.metric.Route
		current.metric.StatusCode = state.metric.StatusCode
		current.metric.RequestsTotal += state.metric.RequestsTotal
		current.metric.LatencySamplesTotal += state.metric.LatencySamplesTotal
		current.metric.LatencyTotalMilliseconds += state.metric.LatencyTotalMilliseconds
		if state.metric.LatencyMaxMilliseconds > current.metric.LatencyMaxMilliseconds {
			current.metric.LatencyMaxMilliseconds = state.metric.LatencyMaxMilliseconds
		}
		if state.lastObservedAt.After(current.lastObservedAt) {
			current.lastObservedAt = state.lastObservedAt
			current.metric.LatencyLastMilliseconds = state.metric.LatencyLastMilliseconds
		}
		byKey[key] = current
	}
}
