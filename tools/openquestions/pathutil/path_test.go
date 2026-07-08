package pathutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRepoRootFindsGoMod(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := FindRepoRoot(nested)
	if err != nil {
		t.Fatal(err)
	}
	if got != root {
		t.Fatalf("root = %q, want %q", got, root)
	}
}

func TestFindRepoRootRejectsMissingRoot(t *testing.T) {
	_, err := FindRepoRoot(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "go.mod not found") {
		t.Fatalf("error = %v, want go.mod not found", err)
	}
}

func TestRepoPath(t *testing.T) {
	if got := RepoPath("/repo", "docs/a.md"); got != filepath.Join("/repo", "docs", "a.md") {
		t.Fatalf("relative path = %q", got)
	}
}
