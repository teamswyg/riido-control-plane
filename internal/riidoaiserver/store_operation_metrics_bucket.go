package riidoaiserver

import "time"

func (m *StoreOperationMetrics) snapshotBuckets(now time.Time) []storeOperationBucket {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneLocked(now)

	buckets := make([]storeOperationBucket, 0, len(m.buckets))
	for _, bucket := range m.buckets {
		copied := storeOperationBucket{
			start:          bucket.start,
			byOperation:    make(map[string]storeOperationMetricState, len(bucket.byOperation)),
			callsTotal:     bucket.callsTotal,
			errorsTotal:    bucket.errorsTotal,
			latency:        bucket.latency,
			lastObservedAt: bucket.lastObservedAt,
		}
		for operation, metric := range bucket.byOperation {
			copied.byOperation[operation] = metric
		}
		buckets = append(buckets, copied)
	}
	return buckets
}

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
