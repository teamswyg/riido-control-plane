package main

import (
	"path/filepath"
	"testing"
)

func TestLoadManifestRejectsMissingAndMalformedInput(t *testing.T) {
	tmp := t.TempDir()
	if _, err := loadManifest(filepath.Join(tmp, "missing.json")); err == nil {
		t.Fatalf("expected missing manifest error")
	}
	for name, body := range map[string]string{"bad-json": `{`, "trailing": `{}` + "\n{}"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(tmp, name+".json")
			writeWebFrontendFile(t, path, body)
			if _, err := loadManifest(path); err == nil {
				t.Fatalf("expected manifest decode error")
			}
		})
	}
}

func TestWriteHelpersReportFailures(t *testing.T) {
	tmp := t.TempDir()
	if err := writeText(filepath.Join(tmp, "ok.md"), "ok"); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(tmp, "ok.json"), map[string]string{"ok": "yes"}); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(tmp, "bad.json"), make(chan int)); err == nil {
		t.Fatalf("expected JSON marshal error")
	}
	blocker := filepath.Join(tmp, "blocker")
	writeWebFrontendFile(t, blocker, "file")
	if err := writeText(filepath.Join(blocker, "x.md"), "x"); err == nil {
		t.Fatalf("expected writeText parent path error")
	}
	if err := writeJSON(filepath.Join(blocker, "x.json"), map[string]string{}); err == nil {
		t.Fatalf("expected writeJSON parent path error")
	}
}
