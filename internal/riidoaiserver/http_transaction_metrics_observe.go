package riidoaiserver

import "net/http"

func NewHTTPTransactionMetrics() *HTTPTransactionMetrics {
	return &HTTPTransactionMetrics{
		buckets: map[int64]*httpTransactionBucket{},
	}
}

func (m *HTTPTransactionMetrics) ObserveHTTPTransaction(obs HTTPTransactionObservation) {
	if m == nil {
		return
	}
	method := normalizeHTTPMetricValue(obs.Method, http.MethodGet)
	route := normalizeHTTPMetricValue(obs.Route, unknownHTTPRoute)
	clientSurface := normalizeHTTPMetricValue(obs.ClientSurface, "unknown")
	statusCode := obs.StatusCode
	if statusCode <= 0 {
		statusCode = http.StatusOK
	}
	observedAt := metricsObservedAt(obs.ObservedAt)
	bucketStart := metricsBucketStart(observedAt)
	elapsedMS := durationMilliseconds(obs.Duration)
	durationKind := httpTransactionDurationKind(obs)
	key := httpTransactionKey{method: method, route: route, clientSurface: clientSurface, statusCode: statusCode, durationKind: durationKind}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneLocked(observedAt)
	bucket := m.bucketLocked(bucketStart)
	bucket.requestsTotal++
	bucket.responsesByStatus[statusCode]++
	if durationKind == "" {
		bucket.latency.observe(elapsedMS)
		if observedAt.After(bucket.lastObservedAt) {
			bucket.lastObservedAt = observedAt
		}
	}
	state := bucket.byKey[key]
	metric := state.metric
	metric.DurationKind = durationKind
	metric.Method = method
	metric.Route = route
	metric.ClientSurface = clientSurface
	metric.StatusCode = statusCode
	metric.RequestsTotal++
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
	bucket.byKey[key] = state
}

func httpTransactionDurationKind(obs HTTPTransactionObservation) string {
	if obs.Stream {
		return "stream_lifetime"
	}
	if obs.Method == http.MethodPost && obs.Route == "/v1/agents/{agent_id}/poll" {
		return "poll_wait"
	}
	return ""
}
