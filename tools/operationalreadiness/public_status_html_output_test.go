package main

import (
	"os"
	"strings"
	"testing"
)

func TestOperationalReadinessWritesPublicStatusHTML(t *testing.T) {
	t.Setenv("GITHUB_WORKFLOW", "public-qa-status-pages")
	t.Setenv("GITHUB_SHA", "abc123")
	t.Setenv("GITHUB_RUN_ID", "456")
	out := t.TempDir() + "/public-status.html"
	err := run(options{
		Repo:             "../..",
		Manifest:         defaultManifest,
		CheckDoc:         true,
		PublicStatusHTML: out,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := readPublicStatusHTML(t, out)
	if !strings.Contains(got, "<title>Riido Public QA Status</title>") {
		t.Fatalf("public status html title missing: %s", got)
	}
	if !strings.Contains(got, "Source workflow:") || !strings.Contains(got, "abc123") {
		t.Fatalf("public status html source missing: %s", got)
	}
	if !strings.Contains(got, "Blocking categories") || !strings.Contains(got, "partial") {
		t.Fatalf("public status html blocking categories missing: %s", got)
	}
	if strings.Contains(got, "eyJ") || strings.Contains(got, "RIIDO_DEVICE_SECRET") {
		t.Fatalf("public status html leaked secret marker: %s", got)
	}
}

func readPublicStatusHTML(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
