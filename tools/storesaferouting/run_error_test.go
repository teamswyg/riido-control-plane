package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMainRunRejectsBadFlags(t *testing.T) {
	t.Parallel()
	if err := mainRun([]string{"-missing"}); err == nil {
		t.Fatal("expected flag parse error")
	}
}

func TestRunRejectsMissingRepoRoot(t *testing.T) {
	t.Parallel()
	err := run(options{Repo: filepath.Join(t.TempDir(), "missing")})
	if err == nil {
		t.Fatal("expected repo root error")
	}
}

func TestRunRejectsMissingManifest(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := writeText(filepath.Join(root, "go.mod"), "module test\n"); err != nil {
		t.Fatal(err)
	}
	err := run(options{Repo: root, Manifest: "missing.json"})
	if err == nil || !strings.Contains(err.Error(), "missing.json") {
		t.Fatalf("error = %v, want missing manifest", err)
	}
}
