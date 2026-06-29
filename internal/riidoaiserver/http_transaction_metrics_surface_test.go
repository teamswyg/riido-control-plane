package riidoaiserver

import (
	"net/http"
	"testing"
	"time"
)

func TestHTTPTransactionMetricsTrackClientSurfaceTotals(t *testing.T) {
	metrics := NewHTTPTransactionMetrics()
	metrics.ObserveHTTPTransaction(HTTPTransactionObservation{
		Method: http.MethodPost, Route: "/v1/agents/{agent_id}/poll",
		ClientSurface: "daemon", StatusCode: http.StatusOK, Duration: time.Millisecond,
	})
	metrics.ObserveHTTPTransaction(HTTPTransactionObservation{
		Method: http.MethodGet, Route: "/v3/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/threads",
		ClientSurface: "client_app", StatusCode: http.StatusOK, Duration: time.Millisecond,
	})
	metrics.ObserveHTTPTransaction(HTTPTransactionObservation{
		Method: http.MethodGet, Route: "/v3/client/workspaces/{workspace_id}/ai-agent/tasks/{task_id}/threads",
		ClientSurface: "client_app", StatusCode: http.StatusOK, Duration: time.Millisecond,
	})
	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if snapshot.HTTPRequestsDaemonTotal != 1 || snapshot.HTTPRequestsClientAppTotal != 2 {
		t.Fatalf("client surface totals = %+v", snapshot)
	}
	if len(snapshot.HTTPTransactions) != 2 || snapshot.HTTPTransactions[0].ClientSurface != "client_app" {
		t.Fatalf("client surface breakdown = %+v", snapshot.HTTPTransactions)
	}
}
