package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadOperationsRejectsInvalidJSON(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "openapi.json")
	if err := os.WriteFile(path, []byte(`{"paths":`), 0o644); err != nil {
		t.Fatalf("write openapi: %v", err)
	}
	if _, err := readOperations(path); err == nil {
		t.Fatal("readOperations succeeded for invalid JSON")
	}
}

func TestFileHashesReturnsReadError(t *testing.T) {
	t.Parallel()
	_, err := fileHashes(map[string]string{"missing": filepath.Join(t.TempDir(), "missing.ts")})
	if err == nil {
		t.Fatal("fileHashes succeeded for missing file")
	}
}

func TestWriteGeneratedFilesWrapsFileName(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(root, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}
	err := writeGeneratedFiles(config{Out: root}, map[string]string{}, nil)
	assertErrorContains(t, err, "write ")
}
