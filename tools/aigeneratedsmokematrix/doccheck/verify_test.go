package doccheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyGeneratedDoc(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.md")
	if err := os.WriteFile(path, []byte("same"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Verify(path, "same"); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyRejectsDrift(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.md")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := Verify(path, "new")
	if err == nil || !strings.Contains(err.Error(), "aigeneratedsmokematrix") {
		t.Fatalf("error = %v, want tool-specific stale message", err)
	}
}
