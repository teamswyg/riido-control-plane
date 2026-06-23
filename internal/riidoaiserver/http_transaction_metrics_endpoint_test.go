package riidoaiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPMetricsIncludesTransactionSnapshot(t *testing.T) {
	store := NewStore()
	defer store.Close()
	metrics := NewHTTPTransactionMetrics()
	server := NewServer(ServerConfig{
		Assignment:       store,
		Authorizer:       assignmentHTTPAuthorizer(t, []string{"metrics:read"}),
		HTTPTransactions: metrics,
	}).Handler()

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
	if snapshot.HTTPRequestsTotal != 0 {
		t.Fatalf("http request total = %d, want metrics endpoint excluded", snapshot.HTTPRequestsTotal)
	}
	if len(snapshot.HTTPTransactions) != 0 {
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
