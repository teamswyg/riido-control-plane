package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readRepositoryManifest(t *testing.T) manifest {
	t.Helper()
	document, err := loadManifest("../..", repositoryContract)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func writeManifest(t *testing.T, document manifest) string {
	t.Helper()
	repo := copyFixtureRepo(t)
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repo, filepath.FromSlash(repositoryContract))
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return repo
}

func copyFixtureRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	paths := []string{
		repositoryContract, ".github/workflows/ci.yml", ".github/workflows/repository-readme.yml",
		"pipelines/control-plane.local-self-check.riido.json", "tools/riido-ci-local", "go.mod",
	}
	for _, path := range paths {
		raw, err := os.ReadFile(filepath.Join("../..", filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(repo, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return repo
}
