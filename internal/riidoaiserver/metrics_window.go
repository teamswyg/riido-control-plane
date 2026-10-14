package riidoaiserver

import "time"

const (
	defaultMetricsWindow  = 5 * time.Minute
	metricsBucketDuration = time.Minute
	metricsBreakdownLimit = 20
)

func metricsObservedAt(observedAt time.Time) time.Time {
	if observedAt.IsZero() {
		return time.Now()
	}
	return observedAt
}

func metricsBucketStart(observedAt time.Time) time.Time {
	return observedAt.Truncate(metricsBucketDuration)
}

func metricBucketExpired(bucketStart, cutoff time.Time) bool {
	return bucketStart.Add(metricsBucketDuration).Before(cutoff)
}
