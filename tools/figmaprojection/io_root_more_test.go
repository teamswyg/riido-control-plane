package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTextAndJSONIORejectBadPaths(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	blocker := filepath.Join(root, "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	if _, err := readText(filepath.Join(root, "missing")); err == nil {
		t.Fatalf("readText should reject missing file")
	}
	if err := writeText(filepath.Join(blocker, "child"), "x"); err == nil {
		t.Fatalf("writeText should reject file parent")
	}
	if err := writeJSON(filepath.Join(blocker, "child"), map[string]string{}); err == nil {
		t.Fatalf("writeJSON should reject file parent")
	}
}

func TestLoadJSONFileRejectsMalformedAndMultipleDocuments(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	bad := filepath.Join(root, "bad.json")
	if err := os.WriteFile(bad, []byte("{"), 0o644); err != nil {
		t.Fatalf("write bad json: %v", err)
	}
	if _, err := loadJSONFile[map[string]string](bad); err == nil {
		t.Fatalf("loadJSONFile should reject malformed JSON")
	}
	two := filepath.Join(root, "two.json")
	if err := os.WriteFile(two, []byte(`{"a":"b"} {}`), 0o644); err != nil {
		t.Fatalf("write two json docs: %v", err)
	}
	if _, err := loadJSONFile[map[string]string](two); err == nil {
		t.Fatalf("loadJSONFile should reject multiple JSON documents")
	}
}

func TestGeneratedDocAndProjectionGateFailures(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := writeDocFile(root, "doc"); err != nil {
		t.Fatalf("writeDocFile: %v", err)
	}
	if err := checkGeneratedDoc(root, "other"); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("checkGeneratedDoc error = %v, want stale", err)
	}
	if err := checkProjectionGate(root); err == nil || !strings.Contains(err.Error(), "failed") {
		t.Fatalf("checkProjectionGate error = %v, want failure", err)
	}
}
