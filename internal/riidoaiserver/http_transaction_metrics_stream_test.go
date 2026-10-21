package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPStreamLifetimeDoesNotInflateRequestLatency(t *testing.T) {
	metrics := NewHTTPTransactionMetrics()
	metrics.ObserveHTTPTransaction(HTTPTransactionObservation{Method: http.MethodGet, Route: "/v2/client/workspaces/{workspace_id}/ai-agent/events", StatusCode: 200, Duration: 12 * time.Minute, Stream: true})
	metrics.ObserveHTTPTransaction(HTTPTransactionObservation{Method: http.MethodGet, Route: "/v3/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/threads", StatusCode: 200, Duration: 25 * time.Millisecond})
	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{})
	if snapshot.HTTPRequestsTotal != 2 || snapshot.HTTPRequestLatencySamplesTotal != 1 || snapshot.HTTPRequestLatencyMaxMilliseconds != 25 {
		t.Fatalf("stream inflated request latency: %+v", snapshot)
	}
	found := false
	for _, metric := range snapshot.HTTPTransactions {
		if metric.DurationKind == "stream_lifetime" {
			found = true
			if metric.LatencyMaxMilliseconds != 720000 {
				t.Fatalf("stream lifetime lost: %+v", metric)
			}
		}
	}
	if !found {
		t.Fatal("missing stream category")
	}
}

func TestHTTPMiddlewareClassifiesSSEFromResponseContentType(t *testing.T) {
	metrics := NewHTTPTransactionMetrics()
	handler := withHTTPTransactionMetrics(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
	}), metrics)
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/ws/ai-agent/events", nil))
	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{})
	if snapshot.HTTPRequestLatencySamplesTotal != 0 || len(snapshot.HTTPTransactions) != 1 || snapshot.HTTPTransactions[0].DurationKind != "stream_lifetime" {
		t.Fatalf("SSE classification=%+v", snapshot)
	}
}
