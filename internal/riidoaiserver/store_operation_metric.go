package riidoaiserver

import (
	"sync"
	"time"
)

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
