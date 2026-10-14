package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigReferenceIOAndPathEdges(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	textPath := filepath.Join(dir, "nested", "doc.md")
	if err := writeText(textPath, "hello"); err != nil {
		t.Fatalf("writeText: %v", err)
	}
	if got, err := os.ReadFile(textPath); err != nil || string(got) != "hello" {
		t.Fatalf("read written text = %q/%v", got, err)
	}
	absolute := filepath.Join(dir, "abs.json")
	if repoPath("/ignored", absolute) != absolute {
		t.Fatal("repoPath must keep absolute paths")
	}
	if _, err := findRepoRoot(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("expected missing go.mod error")
	}
}

func TestConfigReferenceLoadManifestErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	trailing := filepath.Join(dir, "trailing.json")
	if err := os.WriteFile(trailing, []byte(`{"schema_version":"x"} {}`), 0o644); err != nil {
		t.Fatalf("write trailing manifest: %v", err)
	}
	if _, err := loadManifest(trailing); err == nil || !strings.Contains(err.Error(), "one JSON object") {
		t.Fatalf("expected trailing object error, got %v", err)
	}
	if err := writeJSON(filepath.Join(dir, "bad.json"), make(chan int)); err == nil {
		t.Fatal("expected unsupported JSON value error")
	}
}
