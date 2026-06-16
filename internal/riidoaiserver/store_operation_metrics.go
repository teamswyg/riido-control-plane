package riidoaiserver

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

const unknownStoreOperation = "store_unknown"

type StoreOperationName string

const (
	StoreOperationCreateTask       StoreOperationName = "store_create_task"
	StoreOperationCreateAssignment StoreOperationName = "store_create_assignment"
	StoreOperationCancelAssignment StoreOperationName = "store_cancel_assignment"
	StoreOperationPollAssignment   StoreOperationName = "store_poll_assignment"
	StoreOperationWaitAssignment   StoreOperationName = "store_wait_assignment"
	StoreOperationLeaseAssignment  StoreOperationName = "store_lease_assignment"
	StoreOperationAppendEvent      StoreOperationName = "store_append_event"
)

func (op StoreOperationName) String() string {
	value := strings.TrimSpace(string(op))
	if value == "" {
		return unknownStoreOperation
	}
	return value
}

type StoreOperationMetric struct {
	Operation                string `json:"operation"`
	CallsTotal               int64  `json:"calls_total"`
	ErrorsTotal              int64  `json:"errors_total"`
	LatencySamplesTotal      int64  `json:"latency_samples_total"`
	LatencyTotalMilliseconds int64  `json:"latency_total_ms"`
	LatencyMaxMilliseconds   int64  `json:"latency_max_ms"`
	LatencyLastMilliseconds  int64  `json:"latency_last_ms"`
}

type StoreOperationObservation struct {
	Operation  StoreOperationName
	Duration   time.Duration
	Err        error
	ObservedAt time.Time
}

type StoreOperationMetrics struct {
	mu      sync.Mutex
	buckets map[int64]*storeOperationBucket
}

type storeOperationBucket struct {
	start          time.Time
	byOperation    map[string]storeOperationMetricState
	callsTotal     int64
	errorsTotal    int64
	latency        storeOperationLatencyMetrics
	lastObservedAt time.Time
}

type storeOperationMetricState struct {
	metric         StoreOperationMetric
	lastObservedAt time.Time
}

type storeOperationLatencyMetrics struct {
	samplesTotal      int64
	totalMilliseconds int64
	maxMilliseconds   int64
	lastMilliseconds  int64
}

func NewStoreOperationMetrics() *StoreOperationMetrics {
	return &StoreOperationMetrics{
		buckets: map[int64]*storeOperationBucket{},
	}
}

func (m *StoreOperationMetrics) ObserveStoreOperation(obs StoreOperationObservation) {
	if m == nil {
		return
	}
	operation := obs.Operation.String()
	observedAt := metricsObservedAt(obs.ObservedAt)
	bucketStart := metricsBucketStart(observedAt)
	elapsedMS := durationMilliseconds(obs.Duration)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.pruneLocked(observedAt)
	bucket := m.bucketLocked(bucketStart)
	bucket.callsTotal++
	if storeOperationFailed(obs.Err) {
		bucket.errorsTotal++
	}
	bucket.latency.observe(elapsedMS)
	if observedAt.After(bucket.lastObservedAt) {
		bucket.lastObservedAt = observedAt
	}

	state := bucket.byOperation[operation]
	metric := state.metric
	metric.Operation = operation
	metric.CallsTotal++
	if storeOperationFailed(obs.Err) {
		metric.ErrorsTotal++
	}
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
	bucket.byOperation[operation] = state
}

func storeOperationFailed(err error) bool {
	return err != nil && !errors.Is(err, context.Canceled)
}

func (m *StoreOperationMetrics) ApplyToMetricsSnapshot(snapshot MetricsSnapshot) MetricsSnapshot {
	if m == nil {
		return snapshot
	}
	callsTotal, errorsTotal, latency, operations := m.snapshot()
	snapshot.StoreOperationCallsTotal = callsTotal
	snapshot.StoreOperationErrorsTotal = errorsTotal
	snapshot.StoreOperationLatencySamplesTotal = latency.samplesTotal
	snapshot.StoreOperationLatencyTotalMilliseconds = latency.totalMilliseconds
	snapshot.StoreOperationLatencyMaxMilliseconds = latency.maxMilliseconds
	snapshot.StoreOperationLatencyLastMilliseconds = latency.lastMilliseconds
	snapshot.StoreOperations = operations
	return snapshot
}

func (m *StoreOperationMetrics) snapshot() (int64, int64, storeOperationLatencyMetrics, []StoreOperationMetric) {
	buckets := m.snapshotBuckets(time.Now())
	callsTotal := int64(0)
	errorsTotal := int64(0)
	latency := storeOperationLatencyMetrics{}
	latencyLastObservedAt := time.Time{}
	byOperation := map[string]storeOperationMetricState{}
	for _, bucket := range buckets {
		callsTotal += bucket.callsTotal
		errorsTotal += bucket.errorsTotal
		latency.samplesTotal += bucket.latency.samplesTotal
		latency.totalMilliseconds += bucket.latency.totalMilliseconds
		if bucket.latency.maxMilliseconds > latency.maxMilliseconds {
			latency.maxMilliseconds = bucket.latency.maxMilliseconds
		}
		if bucket.lastObservedAt.After(latencyLastObservedAt) {
			latencyLastObservedAt = bucket.lastObservedAt
			latency.lastMilliseconds = bucket.latency.lastMilliseconds
		}
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
	operations := make([]StoreOperationMetric, 0, len(byOperation))
	for _, state := range byOperation {
		operations = append(operations, state.metric)
	}
	sortStoreOperationMetrics(operations)
	if len(operations) > metricsBreakdownLimit {
		operations = operations[:metricsBreakdownLimit]
	}
	return callsTotal, errorsTotal, latency, operations
}

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

func sortStoreOperationMetrics(operations []StoreOperationMetric) {
	sort.Slice(operations, func(i, j int) bool {
		if operations[i].CallsTotal != operations[j].CallsTotal {
			return operations[i].CallsTotal > operations[j].CallsTotal
		}
		return operations[i].Operation < operations[j].Operation
	})
}

func (m *storeOperationLatencyMetrics) observe(elapsedMS int64) {
	m.samplesTotal++
	m.totalMilliseconds += elapsedMS
	if elapsedMS > m.maxMilliseconds {
		m.maxMilliseconds = elapsedMS
	}
	m.lastMilliseconds = elapsedMS
}
