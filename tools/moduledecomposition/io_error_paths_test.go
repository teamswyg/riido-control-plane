package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestRejectsInvalidInputs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := loadManifest(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("expected missing manifest error")
	}
	invalid := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(invalid, []byte(`{"schema_version":"x"} {"extra":true}`), 0o644); err != nil {
		t.Fatalf("write invalid manifest: %v", err)
	}
	if _, err := loadManifest(invalid); err == nil || !strings.Contains(err.Error(), "one JSON object") {
		t.Fatalf("expected one-object error, got %v", err)
	}
	unknown := filepath.Join(dir, "unknown.json")
	if err := os.WriteFile(unknown, []byte(`{"schema_version":"x","unknown":true}`), 0o644); err != nil {
		t.Fatalf("write unknown manifest: %v", err)
	}
	if _, err := loadManifest(unknown); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func TestWriteJSONAndTextErrorPaths(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := writeText(filepath.Join(dir, "nested", "out.txt"), "hello"); err != nil {
		t.Fatalf("write text: %v", err)
	}
	if err := writeJSON(filepath.Join(dir, "bad.json"), make(chan int)); err == nil {
		t.Fatal("expected json marshal error")
	}
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("file"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	if err := writeText(filepath.Join(blocker, "out.txt"), "nope"); err == nil {
		t.Fatal("expected mkdir error through file blocker")
	}
}
