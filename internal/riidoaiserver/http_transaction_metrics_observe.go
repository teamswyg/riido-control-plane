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
	key := httpTransactionKey{method: method, route: route, clientSurface: clientSurface, statusCode: statusCode}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneLocked(observedAt)
	bucket := m.bucketLocked(bucketStart)
	bucket.requestsTotal++
	bucket.responsesByStatus[statusCode]++
	bucket.latency.observe(elapsedMS)
	if observedAt.After(bucket.lastObservedAt) {
		bucket.lastObservedAt = observedAt
	}
	state := bucket.byKey[key]
	metric := state.metric
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
