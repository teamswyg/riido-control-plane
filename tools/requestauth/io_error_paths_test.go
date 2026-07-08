package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestRejectsUnknownFields(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(path, []byte(`{"unknown":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadManifest(path); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("loadManifest error = %v, want unknown field", err)
	}
}

func TestLoadManifestRejectsTrailingJSON(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "manifest.json")
	body := `{"schema_version":"riido-request-authorization.v1"}{}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadManifest(path); err == nil || !strings.Contains(err.Error(), "one JSON value") {
		t.Fatalf("loadManifest error = %v, want one JSON value", err)
	}
}

func TestWriteJSONReportsMarshalAndMkdirErrors(t *testing.T) {
	t.Parallel()
	if err := writeJSON(filepath.Join(t.TempDir(), "out.json"), make(chan int)); err == nil {
		t.Fatal("expected marshal error")
	}
	root := t.TempDir()
	parentFile := filepath.Join(root, "parent")
	if err := os.WriteFile(parentFile, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(parentFile, "out.json"), map[string]string{}); err == nil {
		t.Fatal("expected mkdir error")
	}
}

func TestWriteTextCreatesParentDirectory(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "nested", "out.txt")
	if err := writeText(out, "hello"); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(out); err != nil || string(got) != "hello" {
		t.Fatalf("written text = %q, %v", got, err)
	}
}
