package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLiveWorkflowIOErrorPaths(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := loadManifest(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("expected missing manifest error")
	}
	invalid := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(invalid, []byte("{"), 0o644); err != nil {
		t.Fatalf("write invalid manifest: %v", err)
	}
	if _, err := loadManifest(invalid); err == nil || !strings.Contains(err.Error(), "parse manifest") {
		t.Fatalf("expected parse manifest error, got %v", err)
	}
	if err := writeJSON(filepath.Join(dir, "bad.json"), make(chan int)); err == nil {
		t.Fatal("expected json encode error")
	}
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("file"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	if err := writeText(filepath.Join(blocker, "doc.md"), "x"); err == nil {
		t.Fatal("expected writeText mkdir error")
	}
	if _, err := readText(filepath.Join(dir, "missing.txt")); err == nil {
		t.Fatal("expected readText error")
	}
}

func TestLiveWorkflowPathAndEnvEdges(t *testing.T) {
	dir := t.TempDir()
	if _, err := findRepoRoot(dir); err == nil || !strings.Contains(err.Error(), "go.mod not found") {
		t.Fatalf("expected missing repo root error, got %v", err)
	}
	abs := filepath.Join(dir, "absolute.txt")
	if got := repoPath(dir, abs); got != abs {
		t.Fatalf("repoPath(abs) = %q, want %q", got, abs)
	}
	t.Setenv("RIIDO_LIVE_CHECK_STATUS", "from-env")
	if got := getenvDefault("RIIDO_LIVE_CHECK_STATUS", "fallback"); got != "from-env" {
		t.Fatalf("getenvDefault = %q", got)
	}
}
