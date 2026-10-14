package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaybeDocWritesAndDetectsDrift(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	m := manifest{GeneratedDoc: "docs/audit.md"}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	if err := maybeDoc(root, m, "fresh doc", true, true); err != nil {
		t.Fatalf("maybeDoc write/check: %v", err)
	}
	path := filepath.Join(root, m.GeneratedDoc)
	if err := os.WriteFile(path, []byte("stale doc"), 0o644); err != nil {
		t.Fatalf("write stale doc: %v", err)
	}
	err := maybeDoc(root, m, "fresh doc", false, true)
	assertErrorContains(t, err, "generated doc drift")
}

func TestJSONIOAndPathEdges(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "evidence.json")
	if err := writeJSON(path, map[string]string{"status": "verified"}); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}
	var got map[string]string
	if err := readJSON(path, &got); err != nil || got["status"] != "verified" {
		t.Fatalf("readJSON = %v, %v", got, err)
	}
	if err := writeJSON(path, func() {}); err == nil {
		t.Fatal("writeJSON succeeded for unmarshalable value")
	}
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte(`{"x":`), 0o644); err != nil {
		t.Fatalf("write bad json: %v", err)
	}
	if err := readJSON(bad, &got); err == nil {
		t.Fatal("readJSON succeeded for malformed JSON")
	}
	if abs := repoPath(dir, path); abs != path {
		t.Fatalf("repoPath absolute = %q, want %q", abs, path)
	}
	_, err := findRepoRoot(filepath.Join(dir, "nested"))
	if err == nil || !strings.Contains(err.Error(), "repository root not found") {
		t.Fatalf("findRepoRoot error = %v", err)
	}
}

func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want %q", err, want)
	}
}
