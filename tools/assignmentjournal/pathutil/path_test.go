package pathutil

import (
	"os"
	"path/filepath"
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
	if _, err := FindRepoRoot(t.TempDir()); err == nil {
		t.Fatal("expected missing root error")
	}
}

func TestResolve(t *testing.T) {
	abs := filepath.Join(t.TempDir(), "file")
	if got := Resolve("/repo", abs); got != abs {
		t.Fatalf("absolute path = %q, want %q", got, abs)
	}
	if got := Resolve("/repo", "docs/file.md"); got != filepath.Join("/repo", "docs", "file.md") {
		t.Fatalf("relative path = %q", got)
	}
}
