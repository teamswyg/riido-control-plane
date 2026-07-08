package pathutil

import (
	"path/filepath"
	"testing"
)

func TestResolveKeepsAbsolutePath(t *testing.T) {
	root := t.TempDir()
	abs := filepath.Join(root, "absolute.json")
	if got := Resolve(root, abs); got != abs {
		t.Fatalf("path = %q, want %q", got, abs)
	}
}

func TestResolveConvertsSlashPath(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, "docs", "matrix.json")
	if got := Resolve(root, "docs/matrix.json"); got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}
