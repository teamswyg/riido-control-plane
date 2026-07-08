package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRepoRootRejectsPathWithoutGoMod(t *testing.T) {
	t.Parallel()
	_, err := findRepoRoot(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "go.mod not found") {
		t.Fatalf("findRepoRoot err = %v", err)
	}
}

func TestRepoPathKeepsAbsolutePath(t *testing.T) {
	t.Parallel()
	abs := filepath.Join(t.TempDir(), "file.txt")
	if got := repoPath("/repo", abs); got != abs {
		t.Fatalf("repoPath absolute = %q, want %q", got, abs)
	}
}

func TestWriteJSONRejectsUnmarshalableValue(t *testing.T) {
	t.Parallel()
	err := writeJSON(filepath.Join(t.TempDir(), "out.json"), make(chan int))
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("writeJSON err = %v", err)
	}
}

func TestWriteTextRejectsDirectoryTarget(t *testing.T) {
	t.Parallel()
	err := writeText(t.TempDir(), "body")
	if err == nil {
		t.Fatal("expected writeText directory target error")
	}
}

func TestVerifyDocRejectsMissingGeneratedDoc(t *testing.T) {
	t.Parallel()
	repo := runtimeBoundaryTestRepo(t)
	err := verifyDoc(repo, runtimeBoundaryManifest("missing.md"), "want")
	if err == nil || !strings.Contains(err.Error(), "read generated doc") {
		t.Fatalf("verifyDoc err = %v", err)
	}
}
