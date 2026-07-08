package pathutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepoRootFindsGitDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
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
