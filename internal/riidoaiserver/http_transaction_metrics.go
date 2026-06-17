package riidoaiserver

import (
	"context"
	"net/http"
	"sort"
	"strings"
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
	statusCode := obs.StatusCode
	if statusCode <= 0 {
		statusCode = http.StatusOK
	}
	observedAt := metricsObservedAt(obs.ObservedAt)
	bucketStart := metricsBucketStart(observedAt)
	elapsedMS := durationMilliseconds(obs.Duration)
	key := httpTransactionKey{method: method, route: route, statusCode: statusCode}

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

func (m *HTTPTransactionMetrics) ApplyToMetricsSnapshot(snapshot MetricsSnapshot) MetricsSnapshot {
	if m == nil {
		return snapshot
	}
	requestsTotal, responsesByStatus, latency, transactions := m.snapshot()
	snapshot.HTTPRequestsTotal = requestsTotal
	snapshot.HTTPResponsesByStatus = responsesByStatus
	snapshot.HTTPRequestLatencySamplesTotal = latency.samplesTotal
	snapshot.HTTPRequestLatencyTotalMilliseconds = latency.totalMilliseconds
	snapshot.HTTPRequestLatencyMaxMilliseconds = latency.maxMilliseconds
	snapshot.HTTPRequestLatencyLastMilliseconds = latency.lastMilliseconds
	snapshot.HTTPTransactions = transactions
	return snapshot
}

func (m *HTTPTransactionMetrics) snapshot() (int64, map[int]int64, httpTransactionLatencyMetrics, []HTTPTransactionMetric) {
	buckets := m.snapshotBuckets(time.Now())
	requestsTotal := int64(0)
	responsesByStatus := map[int]int64{}
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
		for key, state := range bucket.byKey {
			current := byKey[key]
			current.metric.Method = state.metric.Method
			current.metric.Route = state.metric.Route
			current.metric.StatusCode = state.metric.StatusCode
			current.metric.RequestsTotal += state.metric.RequestsTotal
			current.metric.LatencySamplesTotal += state.metric.LatencySamplesTotal
			current.metric.LatencyTotalMilliseconds += state.metric.LatencyTotalMilliseconds
			if state.metric.LatencyMaxMilliseconds > current.metric.LatencyMaxMilliseconds {
				current.metric.LatencyMaxMilliseconds = state.metric.LatencyMaxMilliseconds
			}
			if state.lastObservedAt.After(current.lastObservedAt) {
				current.lastObservedAt = state.lastObservedAt
				current.metric.LatencyLastMilliseconds = state.metric.LatencyLastMilliseconds
			}
			byKey[key] = current
		}
	}
	transactions := make([]HTTPTransactionMetric, 0, len(byKey))
	for _, state := range byKey {
		transactions = append(transactions, state.metric)
	}
	sortHTTPTransactionMetrics(transactions)
	if len(transactions) > metricsBreakdownLimit {
		transactions = transactions[:metricsBreakdownLimit]
	}
	return requestsTotal, responsesByStatus, latency, transactions
}

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

func sortHTTPTransactionMetrics(transactions []HTTPTransactionMetric) {
	sort.Slice(transactions, func(i, j int) bool {
		if transactions[i].RequestsTotal != transactions[j].RequestsTotal {
			return transactions[i].RequestsTotal > transactions[j].RequestsTotal
		}
		if transactions[i].Route != transactions[j].Route {
			return transactions[i].Route < transactions[j].Route
		}
		if transactions[i].Method != transactions[j].Method {
			return transactions[i].Method < transactions[j].Method
		}
		return transactions[i].StatusCode < transactions[j].StatusCode
	})
}

func (m *httpTransactionLatencyMetrics) observe(elapsedMS int64) {
	m.samplesTotal++
	m.totalMilliseconds += elapsedMS
	if elapsedMS > m.maxMilliseconds {
		m.maxMilliseconds = elapsedMS
	}
	m.lastMilliseconds = elapsedMS
}

func durationMilliseconds(duration time.Duration) int64 {
	if duration <= 0 {
		return 0
	}
	elapsedMS := duration.Milliseconds()
	if elapsedMS == 0 {
		return 1
	}
	return elapsedMS
}

func normalizeHTTPMetricValue(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func withHTTPTransactionMetrics(next http.Handler, metrics *HTTPTransactionMetrics) http.Handler {
	if metrics == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if excludedHTTPTransactionMetricsPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		startedAt := time.Now()
		recorder := &httpStatusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(recorder, r)
		route := httpMetricRoute(r.Method, r.URL.Path, r.Pattern, recorder.statusCode)
		metrics.ObserveHTTPTransaction(HTTPTransactionObservation{
			Method:     r.Method,
			Route:      route,
			StatusCode: recorder.statusCode,
			Duration:   time.Since(startedAt),
		})
	})
}

func httpMetricRoute(method, path, pattern string, statusCode int) string {
	if route := traceHTTPRoute(method, path); route != "" {
		return route
	}
	if pattern = strings.TrimSpace(pattern); pattern != "" {
		return pattern
	}
	if statusCode == http.StatusNotFound {
		return unmatchedHTTPRoute(path)
	}
	return ""
}

func unmatchedHTTPRoute(path string) string {
	segments := unmatchedHTTPRouteSegments(path)
	if len(segments) == 0 {
		return unmatchedHTTPRoutePrefix + "/"
	}
	first := strings.ToLower(segments[0])
	switch first {
	case "favicon.ico", "robots.txt", "manifest.json", "assetlinks.json", "apple-app-site-association":
		return unmatchedHTTPRoutePrefix + "/" + first
	case ".well-known":
		return unmatchedHTTPRoutePrefix + "/.well-known"
	case "v1", "v2":
		return unmatchedHTTPRoutePrefix + "/" + first + "/" + unmatchedHTTPAPISegment(segments)
	default:
		if strings.Contains(first, ".") {
			return unmatchedHTTPRoutePrefix + "/{asset}"
		}
		return unmatchedHTTPRoutePrefix + "/{other}"
	}
}

func unmatchedHTTPRouteSegments(path string) []string {
	trimmed := strings.Trim(strings.TrimSpace(path), "/")
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, "/")
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			segments = append(segments, part)
		}
	}
	return segments
}

func unmatchedHTTPAPISegment(segments []string) string {
	if len(segments) < 2 {
		return "{other}"
	}
	second := strings.ToLower(segments[1])
	switch second {
	case "agent-catalog", "agents", "client", "component-tasks", "daemon", "desktop":
		return second
	default:
		return "{other}"
	}
}

func excludedHTTPTransactionMetricsPath(path string) bool {
	switch path {
	case "/healthz", "/readyz", "/metrics":
		return true
	default:
		return false
	}
}

type httpStatusRecorder struct {
	http.ResponseWriter
	statusCode int
	wrote      bool
}

func (r *httpStatusRecorder) WriteHeader(statusCode int) {
	if r.wrote {
		return
	}
	r.wrote = true
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *httpStatusRecorder) Write(p []byte) (int, error) {
	if !r.wrote {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(p)
}

func (r *httpStatusRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *httpStatusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

type observedMetricsReader struct {
	base         MetricsReader
	contributors []MetricsSnapshotContributor
}

type MetricsSnapshotContributor interface {
	ApplyToMetricsSnapshot(snapshot MetricsSnapshot) MetricsSnapshot
}

func NewObservedMetricsReader(base MetricsReader, contributors ...MetricsSnapshotContributor) MetricsReader {
	if base == nil {
		return base
	}
	filtered := make([]MetricsSnapshotContributor, 0, len(contributors))
	for _, contributor := range contributors {
		if contributor != nil {
			filtered = append(filtered, contributor)
		}
	}
	if len(filtered) == 0 {
		return base
	}
	return observedMetricsReader{base: base, contributors: filtered}
}

func (r observedMetricsReader) Metrics(ctx context.Context) (MetricsSnapshot, error) {
	snapshot, err := r.base.Metrics(ctx)
	if err != nil {
		return MetricsSnapshot{}, err
	}
	for _, contributor := range r.contributors {
		snapshot = contributor.ApplyToMetricsSnapshot(snapshot)
	}
	return snapshot, nil
}
