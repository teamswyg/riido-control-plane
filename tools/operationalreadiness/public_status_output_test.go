package main

import (
	"os"
	"strings"
	"testing"
)

func TestOperationalReadinessWritesPublicStatusMarkdown(t *testing.T) {
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
	if strings.Contains(got, "eyJ") || strings.Contains(got, "RIIDO_DEVICE_SECRET") {
		t.Fatalf("public status leaked secret marker: %s", got)
	}
}

func TestOperationalReadinessWritesPublicStatusHTML(t *testing.T) {
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
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, "<title>Riido Public QA Status</title>") {
		t.Fatalf("public status html title missing: %s", got)
	}
	if !strings.Contains(got, "Generated at:") || !strings.Contains(got, "Expires at:") {
		t.Fatalf("public status html freshness missing: %s", got)
	}
	if strings.Contains(got, "eyJ") || strings.Contains(got, "RIIDO_DEVICE_SECRET") {
		t.Fatalf("public status html leaked secret marker: %s", got)
	}
}
