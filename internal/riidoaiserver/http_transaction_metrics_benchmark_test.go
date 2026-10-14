package riidoaiserver

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func BenchmarkHTTPTransactionMetricsObserve(b *testing.B) {
	metrics := NewHTTPTransactionMetrics()
	obs := HTTPTransactionObservation{
		Method:     http.MethodPost,
		Route:      "/v2/client/workspaces/{workspace_id}/ai-agent/events",
		StatusCode: http.StatusOK,
		Duration:   2 * time.Millisecond,
	}
	b.ReportAllocs()
	for b.Loop() {
		metrics.ObserveHTTPTransaction(obs)
	}
}

func BenchmarkHTTPTransactionMetricsSnapshot(b *testing.B) {
	metrics := NewHTTPTransactionMetrics()
	for i := range metricsBreakdownLimit * 2 {
		for range i + 1 {
			metrics.ObserveHTTPTransaction(HTTPTransactionObservation{
				Method:     http.MethodPost,
				Route:      fmt.Sprintf("/route-%02d", i),
				StatusCode: http.StatusAccepted,
				Duration:   time.Millisecond,
			})
		}
	}
	b.ReportAllocs()
	for b.Loop() {
		_ = metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	}
}
