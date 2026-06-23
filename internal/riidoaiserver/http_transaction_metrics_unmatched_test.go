package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
