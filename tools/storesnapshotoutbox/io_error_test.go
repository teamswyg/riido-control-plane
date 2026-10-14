package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifestRejectsInvalidInputs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cases := map[string]string{
		"malformed.json": `{"schema_version":`,
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
		t.Fatal("loadManifest accepted missing file")
	}
}

func TestIOHelpersRejectInvalidTargets(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := writeJSON(filepath.Join(root, "evidence.json"), func() {}); err == nil {
		t.Fatal("writeJSON accepted unsupported value")
	}
	blocked := filepath.Join(root, "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeText(filepath.Join(blocked, "doc.md"), "x"); err == nil {
		t.Fatal("writeText accepted file parent")
	}
	if _, err := findRepoRoot(filepath.Join(root, "missing")); err == nil {
		t.Fatal("findRepoRoot accepted path without repo markers")
	}
}

func TestIOHelpersWriteFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	textPath := filepath.Join(root, "doc", "reader.md")
	if err := writeText(textPath, "reader"); err != nil {
		t.Fatalf("writeText: %v", err)
	}
	if got, err := os.ReadFile(textPath); err != nil || string(got) != "reader" {
		t.Fatalf("written text = %q, %v", got, err)
	}
	jsonPath := filepath.Join(root, "evidence", "out.json")
	if err := writeJSON(jsonPath, map[string]string{"status": "verified"}); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}
	if got, err := os.ReadFile(jsonPath); err != nil || !bytes.Contains(got, []byte("verified")) {
		t.Fatalf("written json = %q, %v", got, err)
	}
}
