package riidoaiserver

import (
	"sync"
	"time"
)

const (
	unknownHTTPRoute         = "unknown"
	unmatchedHTTPRoutePrefix = "unmatched:"
)

type HTTPTransactionMetric struct {
	Method                   string `json:"method"`
	Route                    string `json:"route"`
	StatusCode               int    `json:"status_code"`
	RequestsTotal            int64  `json:"requests_total"`
	LatencySamplesTotal      int64  `json:"latency_samples_total"`
	LatencyTotalMilliseconds int64  `json:"latency_total_ms"`
	LatencyMaxMilliseconds   int64  `json:"latency_max_ms"`
	LatencyLastMilliseconds  int64  `json:"latency_last_ms"`
}

type HTTPTransactionObservation struct {
	Method     string
	Route      string
	StatusCode int
	Duration   time.Duration
	ObservedAt time.Time
}

type HTTPTransactionMetrics struct {
	mu      sync.Mutex
	buckets map[int64]*httpTransactionBucket
}

type httpTransactionKey struct {
	method     string
	route      string
	statusCode int
}

type httpTransactionBucket struct {
	start             time.Time
	byKey             map[httpTransactionKey]httpTransactionMetricState
	responsesByStatus map[int]int64
	requestsTotal     int64
	latency           httpTransactionLatencyMetrics
	lastObservedAt    time.Time
}

type httpTransactionMetricState struct {
	metric         HTTPTransactionMetric
	lastObservedAt time.Time
}

type httpTransactionLatencyMetrics struct {
	samplesTotal      int64
	totalMilliseconds int64
	maxMilliseconds   int64
	lastMilliseconds  int64
}
