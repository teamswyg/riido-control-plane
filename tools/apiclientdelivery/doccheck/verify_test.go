package doccheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyGeneratedDoc(t *testing.T) {
	path := filepath.Join(t.TempDir(), "handoff.md")
	if err := os.WriteFile(path, []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Verify(path, "ok"); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyReportsPathForStaleDoc(t *testing.T) {
	path := filepath.Join(t.TempDir(), "handoff.md")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := Verify(path, "new")
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("error = %v, want path in stale message", err)
	}
}
