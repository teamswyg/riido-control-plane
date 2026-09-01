package riidoaiserver

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestHTTPTransactionMetricsLogBoundedServerFailure(t *testing.T) {
	var output bytes.Buffer
	previousWriter := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previousWriter) })

	handler := withHTTPTransactionMetrics(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "private failure", http.StatusServiceUnavailable)
	}), NewHTTPTransactionMetrics())
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/v1/agents/private-agent/poll", nil))

	got := output.String()
	if !strings.Contains(got, `event=http_request_failed`) ||
		!strings.Contains(got, `route="/v1/agents/{agent_id}/poll"`) ||
		strings.Contains(got, "private-agent") || strings.Contains(got, "private failure") {
		t.Fatalf("failure log = %q", got)
	}
}

func TestHTTPTransactionMetricsRecordAndRethrowPanic(t *testing.T) {
	metrics := NewHTTPTransactionMetrics()
	handler := withHTTPTransactionMetrics(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}), metrics)
	func() {
		defer func() {
			if recovered := recover(); recovered != "boom" {
				t.Fatalf("recovered panic = %v", recovered)
			}
		}()
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/poll", nil))
	}()
	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if snapshot.HTTPRequestsTotal != 1 || snapshot.HTTPResponsesByStatus[http.StatusInternalServerError] != 1 {
		t.Fatalf("panic metrics = %+v", snapshot)
	}
}
