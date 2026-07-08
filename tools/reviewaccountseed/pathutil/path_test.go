package pathutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepoRootFindsGoMod(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "nested")
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
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("error = %v, want os.ErrNotExist", err)
	}
}

func TestResolve(t *testing.T) {
	if got := Resolve("/repo", "seed/runtime.json"); got != filepath.Join("/repo", "seed", "runtime.json") {
		t.Fatalf("relative path = %q", got)
	}
}
