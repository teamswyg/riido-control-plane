package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIOHelpersRejectInvalidTargets(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if _, err := readText(filepath.Join(root, "missing.txt")); err == nil {
		t.Fatal("readText accepted missing file")
	}
	if err := writeJSON(filepath.Join(root, "evidence.json"), func() {}); err == nil {
		t.Fatal("writeJSON accepted unsupported value")
	}
	blocked := filepath.Join(root, "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeText(filepath.Join(blocked, "doc.md"), "x"); err == nil {
		t.Fatal("writeText accepted file parent")
	}
	if _, err := findRepoRoot(filepath.Join(root, "missing")); err == nil {
		t.Fatal("findRepoRoot accepted path without repo markers")
	}
}
