package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPTransactionMetricsUseServeMuxPattern(t *testing.T) {
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
	if transaction.Route != "/v1/agents/" || transaction.Method != http.MethodPost || transaction.StatusCode != http.StatusAccepted {
		t.Fatalf("transaction = %+v", transaction)
	}
	if transaction.LatencySamplesTotal != 1 || transaction.LatencyLastMilliseconds <= 0 {
		t.Fatalf("transaction latency = %+v", transaction)
	}
}

func TestHTTPMetricsIncludesTransactionSnapshot(t *testing.T) {
	store := NewStore()
	defer store.Close()
	metrics := NewHTTPTransactionMetrics()
	server := NewServer(ServerConfig{
		Assignment:       store,
		Authorizer:       assignmentHTTPAuthorizer(t, []string{"metrics:read"}),
		HTTPTransactions: metrics,
	}).Handler()

	healthReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	healthResp := httptest.NewRecorder()
	server.ServeHTTP(healthResp, healthReq)
	if healthResp.Code != http.StatusOK {
		t.Fatalf("health status=%d body=%s", healthResp.Code, healthResp.Body.String())
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsReq.Header.Set("Authorization", "Bearer assignment-token")
	metricsResp := httptest.NewRecorder()
	server.ServeHTTP(metricsResp, metricsReq)
	if metricsResp.Code != http.StatusOK {
		t.Fatalf("metrics status=%d body=%s", metricsResp.Code, metricsResp.Body.String())
	}
	var snapshot MetricsSnapshot
	if err := json.Unmarshal(metricsResp.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("metrics json: %v", err)
	}
	if snapshot.HTTPRequestsTotal != 1 {
		t.Fatalf("http request total = %d, want only completed pre-metrics request", snapshot.HTTPRequestsTotal)
	}
	if len(snapshot.HTTPTransactions) != 1 || snapshot.HTTPTransactions[0].Route != "/healthz" {
		t.Fatalf("http transactions = %+v", snapshot.HTTPTransactions)
	}

	after, err := store.Metrics(context.Background())
	if err != nil {
		t.Fatalf("store metrics: %v", err)
	}
	if after.HTTPRequestsTotal != 0 || len(after.HTTPTransactions) != 0 {
		t.Fatalf("store metrics should not own HTTP transactions: %+v", after)
	}
}
