package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRunWritesRedactedLiveEvidence(t *testing.T) {
	server := testServer(t, false)
	var out bytes.Buffer
	err := run([]string{"-base-url", server.URL, "-out", "-"}, &out,
		time.Date(2026, 7, 1, 1, 2, 3, 0, time.UTC), server.Client())
	if err != nil {
		t.Fatal(err)
	}
	var rec record
	if err := json.Unmarshal(out.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	if !rec.Passed || rec.RawBodiesIncluded || rec.SecretsIncluded ||
		rec.StatusVisibility != "public_aggregate" || rec.PagesBuildType != "workflow" {
		t.Fatalf("record = %+v", rec)
	}
}

func TestRunRejectsSecretMarker(t *testing.T) {
	server := testServer(t, true)
	err := run([]string{"-base-url", server.URL}, &bytes.Buffer{}, time.Now(), server.Client())
	if err == nil || !strings.Contains(err.Error(), "secret marker") {
		t.Fatalf("run error = %v", err)
	}
}

func testServer(t *testing.T, secret bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			html := "<html>ok</html>"
			if secret {
				html += " Bearer abcdefghijklmnopqrstuvwxyz"
			}
			_, _ = w.Write([]byte(html))
		case "/status.json":
			_, _ = w.Write([]byte(`{"overall":"degraded","visibility":"public_aggregate","raw_logs_included":false,"secrets_included":false,"endpoint_details":"redacted"}`))
		case "/pages-status.json":
			_, _ = w.Write([]byte(`{"status":"published","visibility":"public_repository","build_type":"workflow","raw_response_included":false,"secrets_included":false}`))
		default:
			http.NotFound(w, r)
		}
	}))
}
