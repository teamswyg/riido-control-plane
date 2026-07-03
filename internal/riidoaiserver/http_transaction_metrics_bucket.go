package riidoaiserver

import "time"

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
