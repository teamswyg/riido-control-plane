package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifestRejectsInvalidInputs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cases := map[string]string{
		"malformed.json": `{"schema_version":`,
		"unknown.json":   `{"schema_version":"x","unknown":true}`,
		"trailing.json":  `{"schema_version":"x"} {"schema_version":"y"}`,
	}
	for name, body := range cases {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := loadManifest(path); err == nil {
			t.Fatalf("loadManifest(%s) succeeded", name)
		}
	}
	if _, err := loadManifest(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("loadManifest missing file succeeded")
	}
}
