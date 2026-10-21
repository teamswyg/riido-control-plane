package riidoaiserver

import (
	"net/http"
	"testing"
	"time"
)

func TestDaemonPollWaitDoesNotInflateInteractiveRequestLatency(t *testing.T) {
	metrics := NewHTTPTransactionMetrics()
	metrics.ObserveHTTPTransaction(HTTPTransactionObservation{Method: http.MethodPost, Route: "/v1/agents/{agent_id}/poll", StatusCode: 200, Duration: 25 * time.Second})
	metrics.ObserveHTTPTransaction(HTTPTransactionObservation{Method: http.MethodGet, Route: "/v3/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/threads", StatusCode: 200, Duration: 25 * time.Millisecond})
	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{})
	if snapshot.HTTPRequestsTotal != 2 || snapshot.HTTPRequestLatencySamplesTotal != 1 || snapshot.HTTPRequestLatencyMaxMilliseconds != 25 {
		t.Fatal("intentional daemon poll wait inflated interactive request latency")
	}
	found := false
	for _, metric := range snapshot.HTTPTransactions {
		if metric.DurationKind == "poll_wait" {
			found = true
			if metric.LatencyMaxMilliseconds != 25000 || metric.LatencySamplesTotal != 1 {
				t.Fatal("poll wait duration was lost")
			}
		}
	}
	if !found {
		t.Fatal("poll wait duration category missing")
	}
}
