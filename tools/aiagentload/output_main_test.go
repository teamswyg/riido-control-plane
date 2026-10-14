package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRunMainWritesReport(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer api.Close()
	out := filepath.Join(t.TempDir(), "report.json")
	err := runMain([]string{
		"-base-url", api.URL,
		"-scenario", "public",
		"-duration", "20ms",
		"-concurrency", "1",
		"-timeout", "1s",
		"-out", out,
	})
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var r report
	if err := json.Unmarshal(body, &r); err != nil {
		t.Fatal(err)
	}
	if r.SchemaVersion != reportSchemaVersion || r.Total == 0 || r.Failures != 0 {
		t.Fatalf("report = %+v", r)
	}
}

func TestWriteReportRejectsUnwritablePath(t *testing.T) {
	err := writeReport(t.TempDir(), report{})
	if err == nil {
		t.Fatal("expected write error")
	}
}
