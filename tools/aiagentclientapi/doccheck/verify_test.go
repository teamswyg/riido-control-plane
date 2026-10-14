package doccheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyAcceptsMatchingGeneratedDoc(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.md")
	if err := os.WriteFile(path, []byte("generated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Verify(path, "generated\n"); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyRejectsStaleGeneratedDoc(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.md")
	if err := os.WriteFile(path, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := Verify(path, "new\n")
	if err == nil || !strings.Contains(err.Error(), "generated doc is stale") {
		t.Fatalf("error = %v, want stale generated doc", err)
	}
}

func TestVerifyWrapsReadError(t *testing.T) {
	err := Verify(filepath.Join(t.TempDir(), "missing.md"), "")
	if err == nil || !strings.Contains(err.Error(), "read generated doc") {
		t.Fatalf("error = %v, want read generated doc", err)
	}
}
