package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
