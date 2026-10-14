package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifestReportsReadAndDecodeErrors(t *testing.T) {
	t.Parallel()
	if _, err := loadManifest(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected read error")
	}
	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(path, []byte(`{`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadManifest(path); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestWriteJSONReportsMarshalAndMkdirErrors(t *testing.T) {
	t.Parallel()
	if err := writeJSON(filepath.Join(t.TempDir(), "out.json"), make(chan int)); err == nil {
		t.Fatal("expected marshal error")
	}
	root := t.TempDir()
	parentFile := filepath.Join(root, "parent")
	if err := os.WriteFile(parentFile, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(parentFile, "out.json"), map[string]string{}); err == nil {
		t.Fatal("expected mkdir error")
	}
}

func TestWriteTextCreatesParentDirectory(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "nested", "out.txt")
	if err := writeText(out, "hello"); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(out); err != nil || string(got) != "hello" {
		t.Fatalf("written text = %q, %v", got, err)
	}
}

func TestWriteTextReportsMkdirError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	parentFile := filepath.Join(root, "parent")
	if err := os.WriteFile(parentFile, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeText(filepath.Join(parentFile, "out.txt"), "hello"); err == nil {
		t.Fatal("expected mkdir error")
	}
}

func TestMainRunRejectsBadFlag(t *testing.T) {
	t.Parallel()
	if err := mainRun([]string{"-nope"}); err == nil {
		t.Fatal("expected flag error")
	}
}
