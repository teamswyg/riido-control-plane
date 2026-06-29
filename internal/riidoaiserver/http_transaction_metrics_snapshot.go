package riidoaiserver

import "time"

func (m *HTTPTransactionMetrics) ApplyToMetricsSnapshot(snapshot MetricsSnapshot) MetricsSnapshot {
	if m == nil {
		return snapshot
	}
	requestsTotal, responsesByStatus, clientSurfaces, latency, transactions := m.snapshot()
	snapshot.HTTPRequestsTotal = requestsTotal
	snapshot.HTTPResponsesByStatus = responsesByStatus
	applyHTTPClientSurfaceTotals(&snapshot, clientSurfaces)
	snapshot.HTTPRequestLatencySamplesTotal = latency.samplesTotal
	snapshot.HTTPRequestLatencyTotalMilliseconds = latency.totalMilliseconds
	snapshot.HTTPRequestLatencyMaxMilliseconds = latency.maxMilliseconds
	snapshot.HTTPRequestLatencyLastMilliseconds = latency.lastMilliseconds
	snapshot.HTTPTransactions = transactions
	return snapshot
}

func (m *HTTPTransactionMetrics) snapshot() (int64, map[int]int64, map[string]int64, httpTransactionLatencyMetrics, []HTTPTransactionMetric) {
	buckets := m.snapshotBuckets(time.Now())
	requestsTotal := int64(0)
	responsesByStatus := map[int]int64{}
	clientSurfaces := map[string]int64{}
	latency := httpTransactionLatencyMetrics{}
	latencyLastObservedAt := time.Time{}
	byKey := map[httpTransactionKey]httpTransactionMetricState{}
	for _, bucket := range buckets {
		requestsTotal += bucket.requestsTotal
		for statusCode, total := range bucket.responsesByStatus {
			responsesByStatus[statusCode] += total
		}
		latency.samplesTotal += bucket.latency.samplesTotal
		latency.totalMilliseconds += bucket.latency.totalMilliseconds
		if bucket.latency.maxMilliseconds > latency.maxMilliseconds {
			latency.maxMilliseconds = bucket.latency.maxMilliseconds
		}
		if bucket.lastObservedAt.After(latencyLastObservedAt) {
			latencyLastObservedAt = bucket.lastObservedAt
			latency.lastMilliseconds = bucket.latency.lastMilliseconds
		}
		mergeHTTPTransactionBucketState(byKey, bucket.byKey)
	}
	transactions := make([]HTTPTransactionMetric, 0, len(byKey))
	for _, state := range byKey {
		clientSurfaces[state.metric.ClientSurface] += state.metric.RequestsTotal
		transactions = append(transactions, state.metric)
	}
	sortHTTPTransactionMetrics(transactions)
	if len(transactions) > metricsBreakdownLimit {
		transactions = transactions[:metricsBreakdownLimit]
	}
	return requestsTotal, responsesByStatus, clientSurfaces, latency, transactions
}
