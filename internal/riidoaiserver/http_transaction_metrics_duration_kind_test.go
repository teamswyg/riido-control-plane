package riidoaiserver

import (
	"net/http"
	"testing"
	"time"
)

func TestHTTPTransactionDurationKindsDoNotMergeForSameRouteAndStatus(t *testing.T) {
	metrics := NewHTTPTransactionMetrics()
	const route = "/v2/client/workspaces/{workspace_id}/ai-agent/events"
	for _, obs := range []HTTPTransactionObservation{
		{Method: http.MethodGet, Route: route, StatusCode: 200, Duration: 5 * time.Millisecond},
		{Method: http.MethodGet, Route: route, StatusCode: 200, Duration: 12 * time.Minute, Stream: true},
	} {
		metrics.ObserveHTTPTransaction(obs)
	}
	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{})
	if len(snapshot.HTTPTransactions) != 2 || snapshot.HTTPTransactions[0].DurationKind != "" || snapshot.HTTPTransactions[1].DurationKind != "stream_lifetime" {
		t.Fatal("same-route request latency and stream lifetime were merged")
	}
	if snapshot.HTTPTransactions[0].LatencyMaxMilliseconds != 5 || snapshot.HTTPTransactions[1].LatencyMaxMilliseconds != 720000 {
		t.Fatal("duration categories contaminated each other's samples")
	}
}
