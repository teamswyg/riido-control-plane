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
	if rec.StatusSourceCommit != "abc123" || rec.PagesSourceRunID != "456" {
		t.Fatalf("record source = %+v", rec)
	}
	if len(rec.StatusBlockingCategories) != 1 ||
		rec.StatusBlockingCategories[0].Category != "usability" {
		t.Fatalf("record blocking categories = %+v", rec.StatusBlockingCategories)
	}
	if rec.BadgeLabel != "riido qa" || !strings.Contains(rec.BadgeMessage, "degraded") {
		t.Fatalf("record badge = %+v", rec)
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
			_, _ = w.Write([]byte(`{"overall":"degraded","visibility":"public_aggregate","source_commit":"abc123","source_run_id":"456","source_run_url":"https://github.com/teamswyg/riido-control-plane/actions/runs/456","raw_logs_included":false,"secrets_included":false,"endpoint_details":"redacted","blocking_categories":[{"category":"usability","partial_count":1,"stale_partial_count":1}]}`))
		case "/pages-status.json":
			_, _ = w.Write([]byte(`{"status":"published","visibility":"public_repository","build_type":"workflow","source_commit":"abc123","source_run_id":"456","source_run_url":"https://github.com/teamswyg/riido-control-plane/actions/runs/456","raw_response_included":false,"secrets_included":false}`))
		case "/status-badge.json":
			_, _ = w.Write([]byte(`{"schemaVersion":1,"label":"riido qa","message":"degraded / 1 categories","color":"orange"}`))
		default:
			http.NotFound(w, r)
		}
	}))
}
