package riidoaiserver

import "time"

func (m *HTTPTransactionMetrics) snapshotBuckets(now time.Time) []httpTransactionBucket {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneLocked(now)
	buckets := make([]httpTransactionBucket, 0, len(m.buckets))
	for _, bucket := range m.buckets {
		copied := httpTransactionBucket{
			start:             bucket.start,
			byKey:             make(map[httpTransactionKey]httpTransactionMetricState, len(bucket.byKey)),
			responsesByStatus: make(map[int]int64, len(bucket.responsesByStatus)),
			requestsTotal:     bucket.requestsTotal,
			latency:           bucket.latency,
			lastObservedAt:    bucket.lastObservedAt,
		}
		for key, metric := range bucket.byKey {
			copied.byKey[key] = metric
		}
		for statusCode, total := range bucket.responsesByStatus {
			copied.responsesByStatus[statusCode] = total
		}
		buckets = append(buckets, copied)
	}
	return buckets
}

func (m *HTTPTransactionMetrics) bucketLocked(start time.Time) *httpTransactionBucket {
	key := start.UnixNano()
	bucket := m.buckets[key]
	if bucket != nil {
		return bucket
	}
	bucket = &httpTransactionBucket{
		start:             start,
		byKey:             map[httpTransactionKey]httpTransactionMetricState{},
		responsesByStatus: map[int]int64{},
	}
	m.buckets[key] = bucket
	return bucket
}

func (m *HTTPTransactionMetrics) pruneLocked(now time.Time) {
	cutoff := now.Add(-defaultMetricsWindow)
	for key, bucket := range m.buckets {
		if metricBucketExpired(bucket.start, cutoff) {
			delete(m.buckets, key)
		}
	}
}
