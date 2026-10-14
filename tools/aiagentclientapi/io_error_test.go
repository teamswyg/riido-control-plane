package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifestRejectsMissingInvalidUnknownAndTrailing(t *testing.T) {
	t.Parallel()
	if _, err := loadManifest(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected missing manifest error")
	}
	for name, body := range map[string]string{
		"invalid.json":  "{",
		"unknown.json":  `{"unknown":true}`,
		"trailing.json": `{} {}`,
	} {
		path := filepath.Join(t.TempDir(), name)
		if err := writeText(path, body); err != nil {
			t.Fatal(err)
		}
		if _, err := loadManifest(path); err == nil {
			t.Fatalf("expected loadManifest(%s) error", name)
		}
	}
}

func TestWriteHelpersRejectBadDirectoriesAndValues(t *testing.T) {
	t.Parallel()
	parentFile := filepath.Join(t.TempDir(), "file")
	if err := writeText(parentFile, "x"); err != nil {
		t.Fatal(err)
	}
	if err := writeText(filepath.Join(parentFile, "child.md"), "x"); err == nil {
		t.Fatal("expected writeText mkdir error")
	}
	if err := writeJSON(filepath.Join(parentFile, "child.json"), map[string]string{}); err == nil {
		t.Fatal("expected writeJSON mkdir error")
	}
	err := writeJSON(filepath.Join(t.TempDir(), "evidence.json"), func() {})
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("writeJSON error = %v", err)
	}
}
