package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTestFunctionExistsRejectsUnsafePackagePath(t *testing.T) {
	tests := []string{"internal/riidoaiserver", "./../secret"}
	for _, pkg := range tests {
		t.Run(pkg, func(t *testing.T) {
			found, err := testFunctionExists(t.TempDir(), pkg, "TestThing")
			if found || err == nil || !strings.Contains(err.Error(), "unsafe package path") {
				t.Fatalf("found=%v err=%v", found, err)
			}
		})
	}
}

func TestLoadManifestRejectsTrailingObject(t *testing.T) {
	path := filepath.Join(t.TempDir(), "risk.json")
	if err := os.WriteFile(path, []byte("{}{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := loadManifest(path)
	if err == nil || !strings.Contains(err.Error(), "single JSON object") {
		t.Fatalf("expected single object error, got %v", err)
	}
}
