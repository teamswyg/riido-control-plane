package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func openQuestionsTempRepo(t *testing.T, manifest string) string {
	t.Helper()
	repo := t.TempDir()
	files := map[string]string{
		"go.mod":          "module tmp\n",
		"open.riido.json": manifest,
	}
	for name, body := range files {
		path := filepath.Join(repo, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return repo
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), want) {
		t.Fatalf("%s does not contain %q:\n%s", path, want, got)
	}
}
