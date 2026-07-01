package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestOperationalReadinessWritesPublicStatusJSON(t *testing.T) {
	t.Setenv("GITHUB_WORKFLOW", "public-qa-status-pages")
	t.Setenv("GITHUB_SHA", "abc123")
	t.Setenv("GITHUB_RUN_ID", "456")
	t.Setenv("GITHUB_SERVER_URL", "https://github.com")
	t.Setenv("GITHUB_REPOSITORY", "teamswyg/riido-control-plane")
	out := t.TempDir() + "/public-status.json"
	err := run(options{
		Repo:             "../..",
		Manifest:         defaultManifest,
		CheckDoc:         true,
		PublicStatusJSON: out,
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var got publicStatus
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Visibility != "public_aggregate" || got.EndpointDetails != "redacted" {
		t.Fatalf("public status json not redacted: %+v", got)
	}
	if got.GeneratedAt == "" || got.ExpiresAt == "" || got.EvidenceTTLHours == 0 {
		t.Fatalf("public status json freshness missing: %+v", got)
	}
	if got.SourceWorkflow != "public-qa-status-pages" || got.SourceCommit != "abc123" {
		t.Fatalf("public status json source missing: %+v", got)
	}
	if got.SourceRunID != "456" || !strings.Contains(got.SourceRunURL, "/actions/runs/456") {
		t.Fatalf("public status json run source missing: %+v", got)
	}
	if strings.Contains(string(data), "eyJ") || strings.Contains(string(data), "RIIDO_DEVICE_SECRET") {
		t.Fatalf("public status json leaked secret marker: %s", data)
	}
}
