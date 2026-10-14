package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestPathHelpersCoverFailureAndAbsolutePath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if _, err := findRepoRoot(root); err == nil {
		t.Fatalf("findRepoRoot should reject directory without .git")
	}
	abs := filepath.Join(root, "file.txt")
	if got := repoPath("/ignored", abs); got != abs {
		t.Fatalf("repoPath absolute = %q, want %q", got, abs)
	}
}

func TestWriteGitHubAnnotationsSkipsNilAndDisabledImpact(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	writeGitHubAnnotations(&out, nil)
	writeGitHubAnnotations(&out, &impactEvidence{})
	if out.Len() != 0 {
		t.Fatalf("annotations should be empty: %q", out.String())
	}
}
