package pathutil

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRepoRootFindsNearestMarker(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	if err := mkdirAll(nested); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filepath.Join(root, "go.mod")); err != nil {
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

func TestFindRepoRootMissingMarker(t *testing.T) {
	_, err := FindRepoRoot(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "repo root not found") {
		t.Fatalf("error = %v, want repo root not found", err)
	}
}

func TestResolve(t *testing.T) {
	root := t.TempDir()
	abs := filepath.Join(root, "abs")
	if got := Resolve(root, abs); got != abs {
		t.Fatalf("absolute path = %q, want %q", got, abs)
	}
	want := filepath.Join(root, "docs", "file.json")
	if got := Resolve(root, "docs/file.json"); got != want {
		t.Fatalf("relative path = %q, want %q", got, want)
	}
}
