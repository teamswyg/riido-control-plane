package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func validPreCommitManifest() manifest {
	return manifest{
		SchemaVersion:   manifestSchema,
		ID:              "pre-commit",
		Title:           "Pre Commit",
		GeneratedDoc:    "docs/pre-commit.md",
		Workflow:        ".github/workflows/precommit.yml",
		Evidence:        "pre-commit-evidence",
		EvidenceTTL:     24,
		PreCommitConfig: ".pre-commit-config.yaml",
		Hooks: []checkBlock{{
			ID: "go-fmt", Summary: "Go fmt", Contains: []string{"id: go-fmt"},
		}},
		Scripts: []scriptSpec{{
			Path: "scripts/check.sh", Summary: "Go test", Contains: []string{"go test ./..."},
		}},
		Loop: completeLoop(),
	}
}

func TestLoadManifestRejectsMissingAndInvalidJSON(t *testing.T) {
	if _, err := loadManifest(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected missing manifest error")
	}
	path := filepath.Join(t.TempDir(), "manifest.json")
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
