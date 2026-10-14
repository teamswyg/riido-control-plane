package smokematrix

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGeneratedPathsTrimsPaths(t *testing.T) {
	path := filepath.Join(t.TempDir(), "matrix.json")
	body := `{"entries":[{"generated_path":" /v2/foo "},{"generated_path":"/v3/bar"}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	paths, count, err := LoadGeneratedPaths(path)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 || len(paths) != 2 || !has(paths, "/v2/foo") || !has(paths, "/v3/bar") {
		t.Fatalf("paths=%v count=%d", paths, count)
	}
}

func TestLoadGeneratedPathsRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "matrix.json")
	if err := os.WriteFile(path, []byte(`{"entries":`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadGeneratedPaths(path); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func has(values map[string]struct{}, key string) bool {
	_, ok := values[key]
	return ok
}
