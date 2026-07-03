package riidoaiserver

import "time"

func (m *StoreOperationMetrics) bucketLocked(start time.Time) *storeOperationBucket {
	key := start.UnixNano()
	bucket := m.buckets[key]
	if bucket != nil {
		return bucket
	}
	bucket = &storeOperationBucket{
		start:       start,
		byOperation: map[string]storeOperationMetricState{},
	}
	m.buckets[key] = bucket
	return bucket
}

func (m *StoreOperationMetrics) pruneLocked(now time.Time) {
	cutoff := now.Add(-defaultMetricsWindow)
	for key, bucket := range m.buckets {
		if metricBucketExpired(bucket.start, cutoff) {
			delete(m.buckets, key)
		}
	}
}
