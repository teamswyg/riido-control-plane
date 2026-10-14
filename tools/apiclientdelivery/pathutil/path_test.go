package pathutil

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRepoRootFindsGoMod(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "docs", "api")
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

func TestFindRepoRootReportsStartPath(t *testing.T) {
	start := t.TempDir()
	_, err := FindRepoRoot(start)
	if err == nil || !strings.Contains(err.Error(), start) {
		t.Fatalf("error = %v, want start path", err)
	}
}

func TestResolve(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, "docs", "handoff.md")
	if got := Resolve(root, "docs/handoff.md"); got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}
