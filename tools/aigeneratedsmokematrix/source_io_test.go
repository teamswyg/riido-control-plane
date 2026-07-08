package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifySourceChecksRejectsMissingOrIncompleteChecks(t *testing.T) {
	if err := verifySourceChecks(t.TempDir(), nil); err == nil {
		t.Fatalf("expected empty source checks error")
	}
	for name, check := range map[string]sourceCheck{
		"incomplete": {Name: "x", File: "source.go"},
		"missing-file": {
			Name:     "x",
			File:     "missing.go",
			Contains: []string{"needle"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := verifySourceCheck(t.TempDir(), check); err == nil {
				t.Fatalf("expected source check error")
			}
		})
	}
}

func TestLoadManifestRejectsMalformedFiles(t *testing.T) {
	tmp := t.TempDir()
	if _, err := loadManifest(filepath.Join(tmp, "missing.json")); err == nil {
		t.Fatalf("expected missing manifest error")
	}
	path := filepath.Join(tmp, "manifest.json")
	if err := os.WriteFile(path, []byte(`{"unknown":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadManifest(path); err == nil {
		t.Fatalf("expected unknown field error")
	}
	if err := os.WriteFile(path, []byte(`{} {}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadManifest(path); err == nil {
		t.Fatalf("expected multi-object manifest error")
	}
}

func TestWritersReportFilesystemAndMarshalErrors(t *testing.T) {
	tmp := t.TempDir()
	if err := writeJSON(filepath.Join(tmp, "ok.json"), map[string]string{"ok": "yes"}); err != nil {
		t.Fatal(err)
	}
	if err := writeText(filepath.Join(tmp, "ok.txt"), "hello"); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(tmp, "bad.json"), make(chan int)); err == nil {
		t.Fatalf("expected JSON marshal error")
	}
	blocker := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blocker, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeText(filepath.Join(blocker, "x.txt"), "x"); err == nil {
		t.Fatalf("expected writeText parent path error")
	}
	if err := writeJSON(filepath.Join(blocker, "x.json"), map[string]string{}); err == nil {
		t.Fatalf("expected writeJSON parent path error")
	}
}
