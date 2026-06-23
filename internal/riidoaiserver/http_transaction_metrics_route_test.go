package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPTransactionMetricsUseRouteVocabulary(t *testing.T) {
	metrics := NewHTTPTransactionMetrics()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agents/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusAccepted, Health{SchemaVersion: SchemaVersion, Status: "ok"})
	})
	handler := withHTTPTransactionMetrics(mux, metrics)

	req := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/poll", nil)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if snapshot.HTTPRequestsTotal != 1 || snapshot.HTTPResponsesByStatus[http.StatusAccepted] != 1 {
		t.Fatalf("http aggregate metrics = %+v", snapshot)
	}
	if len(snapshot.HTTPTransactions) != 1 {
		t.Fatalf("http transaction metrics = %+v", snapshot.HTTPTransactions)
	}
	transaction := snapshot.HTTPTransactions[0]
	if transaction.Route != "/v1/agents/{agent_id}/poll" || transaction.Method != http.MethodPost || transaction.StatusCode != http.StatusAccepted {
		t.Fatalf("transaction = %+v", transaction)
	}
	if transaction.LatencySamplesTotal != 1 || transaction.LatencyLastMilliseconds <= 0 {
		t.Fatalf("transaction latency = %+v", transaction)
	}
}
