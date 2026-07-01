package main

import (
	"os"
	"strings"
	"testing"
)

func TestOperationalReadinessWritesPublicStatusGitHubAnnotation(t *testing.T) {
	out := t.TempDir() + "/public-status.annotation"
	err := run(options{
		Repo:                      "../..",
		Manifest:                  defaultManifest,
		CheckDoc:                  true,
		PublicStatusAnnotationOut: out,
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, "::warning title=Riido Public QA Status::") {
		t.Fatalf("public status annotation missing warning: %s", got)
	}
	if !strings.Contains(got, "generated_at=") || !strings.Contains(got, "expires_at=") {
		t.Fatalf("public status annotation freshness missing: %s", got)
	}
	if strings.Contains(got, "eyJ") || strings.Contains(got, "RIIDO_DEVICE_SECRET") {
		t.Fatalf("public status annotation leaked secret marker: %s", got)
	}
}
