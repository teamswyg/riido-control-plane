package main

import (
	"os"
	"strings"
	"testing"
)

func TestOperationalReadinessWritesPublicStatusMarkdown(t *testing.T) {
	t.Setenv("GITHUB_WORKFLOW", "public-qa-status-pages")
	t.Setenv("GITHUB_SHA", "abc123")
	t.Setenv("GITHUB_RUN_ID", "456")
	out := t.TempDir() + "/public-status.md"
	err := run(options{
		Repo:            "../..",
		Manifest:        defaultManifest,
		CheckDoc:        true,
		PublicStatusOut: out,
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, "## Public QA Status") {
		t.Fatalf("public status heading missing: %s", got)
	}
	if !strings.Contains(got, "- generated at: `") || !strings.Contains(got, "- expires at: `") {
		t.Fatalf("public status freshness missing: %s", got)
	}
	if !strings.Contains(got, "- source workflow: `public-qa-status-pages`") ||
		!strings.Contains(got, "- source commit: `abc123`") {
		t.Fatalf("public status source missing: %s", got)
	}
	if !strings.Contains(got, "- blocking categories:") ||
		!strings.Contains(got, "partial `") {
		t.Fatalf("public status blocking categories missing: %s", got)
	}
	if strings.Contains(got, "eyJ") || strings.Contains(got, "RIIDO_DEVICE_SECRET") {
		t.Fatalf("public status leaked secret marker: %s", got)
	}
}
