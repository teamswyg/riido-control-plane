package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func validManifest() manifest {
	return manifest{
		SchemaVersion: manifestSchema,
		ID:            "go-ci",
		Title:         "Go CI",
		GeneratedDoc:  "docs/go-ci.md",
		Workflow:      ".github/workflows/ci.yml",
		Evidence:      "go-ci-evidence",
		Gates: []gate{{
			ID: "go-test", Summary: "Go tests", Contains: []string{"go test ./..."},
		}},
		Loop: completeLoop(),
	}
}

func TestLoadManifestRejectsMissingAndInvalidJSON(t *testing.T) {
	if _, err := loadManifest(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected missing manifest error")
	}
	root := t.TempDir()
	path := filepath.Join(root, "manifest.json")
	writeFile(t, path, "{")
	if _, err := loadManifest(path); err == nil || !strings.Contains(err.Error(), "decode manifest") {
		t.Fatalf("expected decode manifest error, got %v", err)
	}
}

func TestWriteHelpersReportMarshalAndDirectoryFailures(t *testing.T) {
	if err := writeJSON(filepath.Join(t.TempDir(), "out.json"), make(chan int)); err == nil {
		t.Fatal("expected json marshal error")
	}
	root := t.TempDir()
	blocked := filepath.Join(root, "blocked")
	writeFile(t, blocked, "file")
	if err := writeText(filepath.Join(blocked, "doc.md"), "doc"); err == nil {
		t.Fatal("expected writeText mkdir error")
	}
	if err := writeJSON(filepath.Join(blocked, "out.json"), map[string]string{}); err == nil {
		t.Fatal("expected writeJSON mkdir error")
	}
}
