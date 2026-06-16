package riidoaiserver

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const unknownHTTPRoute = "unknown"

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
}

type HTTPTransactionMetrics struct {
	mu                sync.Mutex
	byKey             map[httpTransactionKey]HTTPTransactionMetric
	responsesByStatus map[int]int64
	requestsTotal     int64
	latency           httpTransactionLatencyMetrics
}

type httpTransactionKey struct {
	method     string
	route      string
	statusCode int
}

type httpTransactionLatencyMetrics struct {
	samplesTotal      int64
	totalMilliseconds int64
	maxMilliseconds   int64
	lastMilliseconds  int64
}

func NewHTTPTransactionMetrics() *HTTPTransactionMetrics {
	return &HTTPTransactionMetrics{
		byKey:             map[httpTransactionKey]HTTPTransactionMetric{},
		responsesByStatus: map[int]int64{},
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
	elapsedMS := durationMilliseconds(obs.Duration)
	key := httpTransactionKey{method: method, route: route, statusCode: statusCode}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestsTotal++
	m.responsesByStatus[statusCode]++
	m.latency.observe(elapsedMS)
	metric := m.byKey[key]
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
	m.byKey[key] = metric
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
	m.mu.Lock()
	defer m.mu.Unlock()
	responsesByStatus := make(map[int]int64, len(m.responsesByStatus))
	for statusCode, total := range m.responsesByStatus {
		responsesByStatus[statusCode] = total
	}
	transactions := make([]HTTPTransactionMetric, 0, len(m.byKey))
	for _, metric := range m.byKey {
		transactions = append(transactions, metric)
	}
	sort.Slice(transactions, func(i, j int) bool {
		if transactions[i].Route != transactions[j].Route {
			return transactions[i].Route < transactions[j].Route
		}
		if transactions[i].Method != transactions[j].Method {
			return transactions[i].Method < transactions[j].Method
		}
		return transactions[i].StatusCode < transactions[j].StatusCode
	})
	return m.requestsTotal, responsesByStatus, m.latency, transactions
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
		startedAt := time.Now()
		recorder := &httpStatusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(recorder, r)
		metrics.ObserveHTTPTransaction(HTTPTransactionObservation{
			Method:     r.Method,
			Route:      r.Pattern,
			StatusCode: recorder.statusCode,
			Duration:   time.Since(startedAt),
		})
	})
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
	transactions *HTTPTransactionMetrics
}

func NewObservedMetricsReader(base MetricsReader, transactions *HTTPTransactionMetrics) MetricsReader {
	if base == nil || transactions == nil {
		return base
	}
	return observedMetricsReader{base: base, transactions: transactions}
}

func (r observedMetricsReader) Metrics(ctx context.Context) (MetricsSnapshot, error) {
	snapshot, err := r.base.Metrics(ctx)
	if err != nil {
		return MetricsSnapshot{}, err
	}
	return r.transactions.ApplyToMetricsSnapshot(snapshot), nil
}
