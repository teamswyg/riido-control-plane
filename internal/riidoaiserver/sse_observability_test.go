package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSSEObservabilitySeparatesTTFBFromStreamLifetime(t *testing.T) {
	metrics := NewHTTPTransactionMetrics()
	trace := &recordingTraceRecorder{}
	streamOpened := make(chan struct{})
	releaseStream := make(chan struct{})
	requestDone := make(chan struct{})

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/client/workspaces/", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Millisecond)
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.(http.Flusher).Flush()
		close(streamOpened)
		<-releaseStream
	})
	handler := withHTTPTracing(withHTTPTransactionMetrics(mux, metrics), trace)
	go func() {
		defer close(requestDone)
		handler.ServeHTTP(
			httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, "/v2/client/workspaces/workspace-sensitive/ai-agent/events", nil),
		)
	}()

	select {
	case <-streamOpened:
	case <-time.After(time.Second):
		t.Fatal("SSE stream did not open")
	}
	time.Sleep(30 * time.Millisecond)

	openSnapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if openSnapshot.HTTPRequestsTotal != 1 || openSnapshot.HTTPRequestLatencySamplesTotal != 1 {
		t.Fatalf("SSE TTFB was not recorded as one HTTP response: %+v", openSnapshot)
	}
	if openSnapshot.SSEStreamsActive != 1 || openSnapshot.SSEStreamsOpenedTotal != 1 ||
		openSnapshot.SSEStreamsClosedTotal != 0 || openSnapshot.SSEStreamDurationSamplesTotal != 0 {
		t.Fatalf("open SSE metrics = %+v", openSnapshot)
	}
	if openSnapshot.SSEStreamTTFBSamplesTotal != 1 || len(openSnapshot.SSEStreams) != 1 {
		t.Fatalf("SSE TTFB metrics = %+v", openSnapshot)
	}
	streamMetric := openSnapshot.SSEStreams[0]
	if streamMetric.Route != "/v2/client/workspaces/{workspace_id}/ai-agent/events" || streamMetric.ActiveStreams != 1 {
		t.Fatalf("SSE route metric = %+v", streamMetric)
	}

	spans := trace.snapshot()
	if len(spans) != 1 || !spans[0].Ended || spans[0].Attributes[riidoHTTPStreamTraceKey] != "sse" ||
		spans[0].Attributes[riidoHTTPTimeToFirstByteTraceKey] == "" {
		t.Fatalf("SSE setup span = %+v", spans)
	}

	close(releaseStream)
	select {
	case <-requestDone:
	case <-time.After(time.Second):
		t.Fatal("SSE request did not close")
	}
	closedSnapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if closedSnapshot.SSEStreamsActive != 0 || closedSnapshot.SSEStreamsClosedTotal != 1 ||
		closedSnapshot.SSEStreamDurationSamplesTotal != 1 || closedSnapshot.SSEStreamDurationMaxMilliseconds < 20 {
		t.Fatalf("closed SSE metrics = %+v", closedSnapshot)
	}
	if closedSnapshot.HTTPRequestLatencyMaxMilliseconds >= closedSnapshot.SSEStreamDurationMaxMilliseconds {
		t.Fatalf("HTTP latency includes stream lifetime: http=%dms stream=%dms",
			closedSnapshot.HTTPRequestLatencyMaxMilliseconds,
			closedSnapshot.SSEStreamDurationMaxMilliseconds,
		)
	}
}

func TestSSEEndpointErrorStaysInRegularHTTPObservability(t *testing.T) {
	metrics := NewHTTPTransactionMetrics()
	trace := &recordingTraceRecorder{}
	handler := withHTTPTracing(withHTTPTransactionMetrics(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Millisecond)
		writeError(w, http.StatusUnauthorized, "unauthorized")
	}), metrics), trace)

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/v1/client/ai-agent/events", nil))
	snapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if resp.Code != http.StatusUnauthorized || snapshot.HTTPResponsesByStatus[http.StatusUnauthorized] != 1 ||
		snapshot.HTTPRequestLatencySamplesTotal != 1 || snapshot.SSEStreamsOpenedTotal != 0 {
		t.Fatalf("failed SSE endpoint metrics = status:%d snapshot:%+v", resp.Code, snapshot)
	}
	spans := trace.snapshot()
	if len(spans) != 1 || !spans[0].Ended || spans[0].Attributes[riidoHTTPStreamTraceKey] != "" {
		t.Fatalf("failed SSE endpoint span = %+v", spans)
	}
}

func TestSSEActiveGaugeSurvivesRollingWindowAndLongLifetimeStaysSeparate(t *testing.T) {
	metrics := NewHTTPTransactionMetrics()
	openedAt := time.Now().Add(-62 * time.Hour)
	metrics.ObserveSSEStreamOpen(SSEStreamOpenObservation{
		Route:           "/v1/client/ai-agent/events",
		ClientSurface:   "client_app",
		TimeToFirstByte: 25 * time.Millisecond,
		ObservedAt:      openedAt,
	})

	openSnapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if openSnapshot.SSEStreamsActive != 1 || openSnapshot.SSEStreamsOpenedTotal != 0 ||
		openSnapshot.SSEStreamTTFBSamplesTotal != 0 {
		t.Fatalf("long-lived active SSE metrics = %+v", openSnapshot)
	}

	metrics.ObserveSSEStreamClose(SSEStreamCloseObservation{
		Route:         "/v1/client/ai-agent/events",
		ClientSurface: "client_app",
		Duration:      62 * time.Hour,
	})
	closedSnapshot := metrics.ApplyToMetricsSnapshot(MetricsSnapshot{SchemaVersion: MetricsSchemaVersion})
	if closedSnapshot.SSEStreamsActive != 0 || closedSnapshot.SSEStreamsClosedTotal != 1 ||
		closedSnapshot.SSEStreamDurationMaxMilliseconds != (62*time.Hour).Milliseconds() ||
		closedSnapshot.HTTPRequestLatencySamplesTotal != 0 {
		t.Fatalf("62-hour SSE lifetime separation = %+v", closedSnapshot)
	}
}
