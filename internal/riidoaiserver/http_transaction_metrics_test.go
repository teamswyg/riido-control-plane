package riidoaiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestHTTPTransactionMetricsClassifiesUnmatchedRoutesWithoutRawPath(t *testing.T) {
	metrics := NewHTTPTransactionMetrics()
	handler := withHTTPTransactionMetrics(http.NewServeMux(), metrics)
	for _, req := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/favicon.ico", nil),
		httptest.NewRequest(http.MethodGet, "/v2/not-a-real-thing/workspace-sensitive", nil),
		httptest.NewRequest(http.MethodGet, "/opaque-sensitive-token-1234567890", nil),
		httptest.NewRequest(http.MethodGet, "/apple-touch-icon.png", nil),
	} {
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code != http.StatusNotFound {
			t.Fatalf("%s status=%d body=%s", req.URL.Path, resp.Code, resp.Body.String())
		}
	}

	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	routes := map[string]bool{}
	for _, transaction := range snapshot.HTTPTransactions {
		routes[transaction.Route] = true
		if strings.Contains(transaction.Route, "sensitive") ||
			strings.Contains(transaction.Route, "opaque") ||
			strings.Contains(transaction.Route, "apple-touch-icon") {
			t.Fatalf("unmatched route should not expose raw path: %+v", transaction)
		}
	}
	for _, want := range []string{
		"unmatched:/favicon.ico",
		"unmatched:/v2/{other}",
		"unmatched:/{other}",
		"unmatched:/{asset}",
	} {
		if !routes[want] {
			t.Fatalf("missing unmatched route %q in %+v", want, snapshot.HTTPTransactions)
		}
	}
	if snapshot.HTTPRequestsTotal != 4 || snapshot.HTTPResponsesByStatus[http.StatusNotFound] != 4 {
		t.Fatalf("http aggregate metrics = %+v", snapshot)
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

func TestHTTPMetricsExcludeOperationalEndpoints(t *testing.T) {
	metrics := NewHTTPTransactionMetrics()
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, Health{SchemaVersion: SchemaVersion, Status: "ok"})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, Health{SchemaVersion: SchemaVersion, Status: "ready"})
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	})
	handler := withHTTPTransactionMetrics(mux, metrics)

	for _, path := range []string{"/healthz", "/readyz", "/metrics"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", path, resp.Code, resp.Body.String())
		}
	}

	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if snapshot.HTTPRequestsTotal != 0 || len(snapshot.HTTPTransactions) != 0 {
		t.Fatalf("operational endpoint metrics should be excluded: %+v", snapshot)
	}
}

func TestHTTPTransactionMetricsUseRollingWindowAndTopBreakdown(t *testing.T) {
	metrics := NewHTTPTransactionMetrics()
	old := time.Now().Add(-10 * time.Minute)
	metrics.ObserveHTTPTransaction(HTTPTransactionObservation{
		Method:     http.MethodGet,
		Route:      "/old",
		StatusCode: http.StatusOK,
		Duration:   time.Millisecond,
		ObservedAt: old,
	})
	for i := range metricsBreakdownLimit + 5 {
		for range i + 1 {
			metrics.ObserveHTTPTransaction(HTTPTransactionObservation{
				Method:     http.MethodPost,
				Route:      fmt.Sprintf("/route-%02d", i),
				StatusCode: http.StatusAccepted,
				Duration:   time.Millisecond,
			})
		}
	}

	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if snapshot.HTTPRequestsTotal <= int64(metricsBreakdownLimit) {
		t.Fatalf("http request total = %d", snapshot.HTTPRequestsTotal)
	}
	if len(snapshot.HTTPTransactions) != metricsBreakdownLimit {
		t.Fatalf("http transaction breakdown size = %d, want %d", len(snapshot.HTTPTransactions), metricsBreakdownLimit)
	}
	if snapshot.HTTPTransactions[0].Route != "/route-24" || snapshot.HTTPTransactions[0].RequestsTotal != 25 {
		t.Fatalf("top transaction = %+v", snapshot.HTTPTransactions[0])
	}
	for _, transaction := range snapshot.HTTPTransactions {
		if transaction.Route == "/old" {
			t.Fatalf("old transaction was not evicted: %+v", snapshot.HTTPTransactions)
		}
	}
}
