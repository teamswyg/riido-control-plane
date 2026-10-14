package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRunAttachesOptInPprofEvidence(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer api.Close()
	pprof := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("pprof-sample"))
	}))
	defer pprof.Close()
	report, err := run(context.Background(), config{
		BaseURL:             api.URL,
		Scenario:            "public",
		Duration:            20 * time.Millisecond,
		Concurrency:         1,
		Timeout:             time.Second,
		PprofBaseURL:        pprof.URL,
		PprofProfileSeconds: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Total == 0 || report.Failures != 0 {
		t.Fatalf("load report = total %d failures %d", report.Total, report.Failures)
	}
	if !report.Pprof.Enabled || len(report.Pprof.Samples) != 3 {
		t.Fatalf("pprof report = %+v", report.Pprof)
	}
}
