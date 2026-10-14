package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIOHelpersRejectInvalidSources(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	badGo := filepath.Join(root, "bad.go")
	if err := os.WriteFile(badGo, []byte("package bad\nimport \""), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := importsFromFile(badGo); err == nil {
		t.Fatal("importsFromFile accepted invalid Go")
	}
	if err := writeJSON(filepath.Join(root, "evidence.json"), func() {}); err == nil {
		t.Fatal("writeJSON accepted unsupported value")
	}
	if err := verifyDoc(root, validTestManifest()); err == nil {
		t.Fatal("verifyDoc accepted missing generated doc")
	}
}
