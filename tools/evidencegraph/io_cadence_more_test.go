package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIORejectsBadParentsAndMalformedManifest(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	blocker := filepath.Join(root, "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	if err := writeText(filepath.Join(blocker, "child"), "x"); err == nil {
		t.Fatalf("writeText should reject file parent")
	}
	if err := writeJSON(filepath.Join(blocker, "child"), map[string]string{}); err == nil {
		t.Fatalf("writeJSON should reject file parent")
	}
	if _, err := loadManifest(filepath.Join(root, "missing.json")); err == nil {
		t.Fatalf("loadManifest should reject missing file")
	}
	bad := filepath.Join(root, "bad.json")
	if err := os.WriteFile(bad, []byte("{"), 0o644); err != nil {
		t.Fatalf("write bad json: %v", err)
	}
	if _, err := loadManifest(bad); err == nil {
		t.Fatalf("loadManifest should reject malformed JSON")
	}
}

func TestPositiveStepMinutes(t *testing.T) {
	t.Parallel()
	got, err := positiveStepMinutes("*/5 * * * *", "*/5", 60)
	if err != nil {
		t.Fatalf("positiveStepMinutes: %v", err)
	}
	if got != 300 {
		t.Fatalf("minutes = %d", got)
	}
	for _, field := range []string{"*", "*/0", "*/bad"} {
		if _, err := positiveStepMinutes("expr", field, 1); err == nil {
			t.Fatalf("field %q should fail", field)
		}
	}
}
