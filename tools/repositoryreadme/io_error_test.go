package main

import (
	"path/filepath"
	"testing"
)

func TestLoadManifestRejectsMalformedUnknownAndTrailingJSON(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"bad.json":      "{",
		"unknown.json":  `{"schema_version":"` + manifestSchema + `","unknown":true}`,
		"trailing.json": `{}` + "\n{}",
	} {
		path := filepath.Join(dir, name)
		writeReadmeTestFile(t, path, body)
		if _, err := loadManifest(path); err == nil {
			t.Fatalf("expected loadManifest error for %s", name)
		}
	}
}

func TestLoadFragmentRejectsMalformedIdentityAndTrailingJSON(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"bad.json":      "{",
		"identity.json": `{"schema_version":"wrong","id":""}`,
		"trailing.json": `{"schema_version":"` + fragmentSchema + `","id":"f"}` + "\n{}",
	} {
		path := filepath.Join(dir, name)
		writeReadmeTestFile(t, path, body)
		if _, err := loadFragment(path); err == nil {
			t.Fatalf("expected loadFragment error for %s", name)
		}
	}
}

func TestWriteTextAndJSONSurfaceErrors(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "parent")
	writeReadmeTestFile(t, parent, "not a dir")
	if err := writeText(filepath.Join(parent, "README.md"), "body"); err == nil {
		t.Fatal("expected text mkdir error")
	}
	if err := writeJSON(filepath.Join(parent, "out.json"), map[string]string{}); err == nil {
		t.Fatal("expected json mkdir error")
	}
	if err := writeJSON(filepath.Join(t.TempDir(), "out.json"), func() {}); err == nil {
		t.Fatal("expected json marshal error")
	}
}
