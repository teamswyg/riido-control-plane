package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPprofCaptureRecordsRedactedSamples(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("pprof-sample"))
	}))
	defer server.Close()
	cfg := config{PprofBaseURL: server.URL, Timeout: time.Second, PprofProfileSeconds: 0}
	got := startPprofCapture(context.Background(), cfg)()
	if !got.Enabled || got.BaseHost == "" || len(got.Samples) != 3 {
		t.Fatalf("pprof evidence = %+v", got)
	}
	for _, sample := range got.Samples {
		if sample.Status != http.StatusOK || sample.Bytes == 0 || sample.ErrorCategory != "" {
			t.Fatalf("pprof sample = %+v", sample)
		}
	}
}

func TestPprofCaptureDisabledByDefault(t *testing.T) {
	got := startPprofCapture(context.Background(), config{})()
	if got.Enabled || len(got.Samples) != 0 {
		t.Fatalf("pprof should be disabled by default: %+v", got)
	}
}
