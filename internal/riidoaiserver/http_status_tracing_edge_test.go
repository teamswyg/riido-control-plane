package riidoaiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPStatusRecorderFlushUnwrapAndImplicitStatus(t *testing.T) {
	base := httptest.NewRecorder()
	recorder := &httpStatusRecorder{ResponseWriter: base}
	if _, err := recorder.Write([]byte("ok")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	recorder.WriteHeader(http.StatusAccepted)
	recorder.Flush()
	if base.Code != http.StatusOK || !base.Flushed {
		t.Fatalf("recorder status/flushed = %d/%v", base.Code, base.Flushed)
	}
	if recorder.Unwrap() != base {
		t.Fatalf("Unwrap did not return base response writer")
	}
}

func TestHTTPTracingRecordsServerErrorStatus(t *testing.T) {
	trace := &recordingTraceRecorder{}
	handler := withHTTPTracing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusServiceUnavailable)
	}), trace)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/poll", nil))
	spans := trace.snapshot()
	if len(spans) != 1 || len(spans[0].Errors) != 1 {
		t.Fatalf("server error span = %+v", spans)
	}
	if got := spans[0].Errors[0]; got != "http status 503" {
		t.Fatalf("trace error = %q, want http status 503", got)
	}
}
