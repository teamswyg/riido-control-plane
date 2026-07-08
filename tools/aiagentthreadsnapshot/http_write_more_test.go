package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetJSONRejectsNonSuccessAndBadJSON(t *testing.T) {
	t.Parallel()
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer errorServer.Close()
	var payload threadCollection
	status, err := getJSON(context.Background(), errorServer.Client(), "token", errorServer.URL, &payload)
	if status != http.StatusServiceUnavailable || err == nil {
		t.Fatalf("expected non-success status error, got %d/%v", status, err)
	}
	badJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{"))
	}))
	defer badJSON.Close()
	status, err = getJSON(context.Background(), badJSON.Client(), "token", badJSON.URL, &payload)
	if status != http.StatusOK || err == nil {
		t.Fatalf("expected decode error, got %d/%v", status, err)
	}
}

func TestEndpointResultCapturesErrorText(t *testing.T) {
	t.Parallel()
	got := endpointResult("v3", "/threads", 400, context.Canceled)
	if got.Method != http.MethodGet || !strings.Contains(got.Error, "canceled") {
		t.Fatalf("endpoint observation lost method/error: %+v", got)
	}
}

func TestWriteReportRejectsBlockedParent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("file"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	if err := writeReport(filepath.Join(blocker, "report.json"), report{}); err == nil {
		t.Fatal("expected blocked parent write error")
	}
}

func TestRunMainRejectsMissingRequiredFlags(t *testing.T) {
	t.Parallel()
	if err := runMain([]string{"-workspace-id", "ws"}); err == nil {
		t.Fatal("expected missing task/output error")
	}
}
