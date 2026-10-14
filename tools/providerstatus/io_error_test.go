package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProviderStatusLoadManifestRejectsMissingMalformedAndExtraJSON(t *testing.T) {
	if _, err := loadManifest(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected missing manifest error")
	}
	path := filepath.Join(t.TempDir(), "manifest.json")
	writeProviderStatusTestFile(t, path, "{")
	if _, err := loadManifest(path); err == nil {
		t.Fatal("expected malformed manifest error")
	}
	writeProviderStatusTestFile(t, path, `{"schema_version":"x"} {}`)
	if _, err := loadManifest(path); err == nil {
		t.Fatal("expected extra JSON object error")
	}
}

func TestProviderStatusWriteJSONSurfacesMarshalAndMkdirErrors(t *testing.T) {
	if err := writeJSON(filepath.Join(t.TempDir(), "out.json"), func() {}); err == nil {
		t.Fatal("expected marshal error")
	}
	parent := filepath.Join(t.TempDir(), "parent")
	writeProviderStatusTestFile(t, parent, "not a dir")
	if err := writeJSON(filepath.Join(parent, "out.json"), map[string]string{}); err == nil {
		t.Fatal("expected mkdir error")
	}
}

func TestProviderStatusWriteTextWritesBodyAndSurfacesMkdirError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.md")
	if err := writeText(path, "hello"); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "hello" {
		t.Fatalf("body = %q", body)
	}
	parent := filepath.Join(t.TempDir(), "parent")
	writeProviderStatusTestFile(t, parent, "not a dir")
	if err := writeText(filepath.Join(parent, "doc.md"), "x"); err == nil {
		t.Fatal("expected mkdir error")
	}
}
