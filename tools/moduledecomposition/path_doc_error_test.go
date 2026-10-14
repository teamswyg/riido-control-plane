package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathAndDocErrorPaths(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := findRepoRoot(dir); err == nil || !strings.Contains(err.Error(), "go.mod not found") {
		t.Fatalf("expected missing repo root error, got %v", err)
	}
	abs := filepath.Join(dir, "doc.md")
	if got := repoPath(dir, abs); got != abs {
		t.Fatalf("absolute repoPath = %q, want %q", got, abs)
	}
	m := validShapeManifest()
	m.GeneratedDoc = "docs/out.md"
	if err := verifyDoc(dir, m, "fresh"); err == nil {
		t.Fatal("expected missing generated doc error")
	}
	path := repoPath(dir, m.GeneratedDoc)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir doc: %v", err)
	}
	if err := os.WriteFile(path, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale doc: %v", err)
	}
	if err := verifyDoc(dir, m, "fresh"); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("expected stale doc error, got %v", err)
	}
}

func TestCountLinesAndNoTargetBudget(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty.go")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatalf("write empty: %v", err)
	}
	if got, err := countLines(empty); err != nil || got != 0 {
		t.Fatalf("empty count = %d, %v", got, err)
	}
	code := filepath.Join(dir, "two.go")
	if err := os.WriteFile(code, []byte("package main\nfunc main() {}"), 0o644); err != nil {
		t.Fatalf("write code: %v", err)
	}
	if got, err := countLines(code); err != nil || got != 2 {
		t.Fatalf("line count = %d, %v", got, err)
	}
	if _, err := countLines(filepath.Join(dir, "missing.go")); err == nil {
		t.Fatal("expected read error")
	}
}
