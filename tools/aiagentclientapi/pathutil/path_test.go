package pathutil

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRepoRootFindsGoMod(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "tools", "api")
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

func TestFindRepoRootMissingGoMod(t *testing.T) {
	_, err := FindRepoRoot(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "go.mod not found") {
		t.Fatalf("error = %v, want go.mod not found", err)
	}
}

func TestResolve(t *testing.T) {
	root := t.TempDir()
	abs := filepath.Join(root, "absolute")
	if got := Resolve(root, abs); got != abs {
		t.Fatalf("absolute path = %q, want %q", got, abs)
	}
	want := filepath.Join(root, "contracts", "client.json")
	if got := Resolve(root, "contracts/client.json"); got != want {
		t.Fatalf("relative path = %q, want %q", got, want)
	}
}
