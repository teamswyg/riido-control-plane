package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestRejectsMissingAndInvalidJSON(t *testing.T) {
	t.Parallel()
	if _, err := loadManifest(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected missing manifest error")
	}
	invalid := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeText(invalid, "{"); err != nil {
		t.Fatal(err)
	}
	if _, err := loadManifest(invalid); err == nil {
		t.Fatal("expected invalid manifest error")
	}
}

func TestWriteJSONRejectsUnmarshalableValues(t *testing.T) {
	t.Parallel()
	err := writeJSON(filepath.Join(t.TempDir(), "evidence.json"), func() {})
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("writeJSON error = %v", err)
	}
}

func TestWriteHelpersRejectFileAsDirectory(t *testing.T) {
	t.Parallel()
	parentFile := filepath.Join(t.TempDir(), "file")
	if err := writeText(parentFile, "x"); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(parentFile, "child.txt")
	if err := writeText(nested, "body"); err == nil {
		t.Fatal("expected writeText mkdir error")
	}
	if err := writeJSON(filepath.Join(parentFile, "child.json"), map[string]string{}); err == nil {
		t.Fatal("expected writeJSON mkdir error")
	}
}
