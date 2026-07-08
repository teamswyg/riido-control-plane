package sourcechecks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resolve(root, file string) string {
	return filepath.Join(root, file)
}

func TestVerifyAcceptsRequiredNeedles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("trace metric status"), 0o644); err != nil {
		t.Fatal(err)
	}
	checks := []Check{{Name: "metrics", File: "source.go", Contains: []string{"trace", "status"}}}
	if err := Verify(root, checks, resolve); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyRejectsMissingNeedle(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("trace"), 0o644); err != nil {
		t.Fatal(err)
	}
	checks := []Check{{Name: "metrics", File: "source.go", Contains: []string{"status"}}}
	err := Verify(root, checks, resolve)
	if err == nil || !strings.Contains(err.Error(), `missing "status"`) {
		t.Fatalf("error = %v, want missing status", err)
	}
}

func TestVerifyWrapsReadError(t *testing.T) {
	err := Verify(t.TempDir(), []Check{{Name: "missing", File: "missing.go"}}, resolve)
	if err == nil || !strings.Contains(err.Error(), "read source check missing") {
		t.Fatalf("error = %v, want read source check", err)
	}
}
