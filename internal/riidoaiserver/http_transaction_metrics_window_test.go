package riidoaiserver

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

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
