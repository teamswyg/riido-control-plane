package main

import "testing"

func loadProjectionFixture(t *testing.T) projectionManifest {
	t.Helper()
	root, err := findRepoRoot(".")
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	value, err := loadJSONFile[projectionManifest](repoPath(root, defaultProjection))
	if err != nil {
		t.Fatalf("load projection: %v", err)
	}
	return value
}

func loadSourceFixture(t *testing.T) sourceManifest {
	t.Helper()
	root, err := findRepoRoot(".")
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	value, err := loadJSONFile[sourceManifest](repoPath(root, defaultSource))
	if err != nil {
		t.Fatalf("load source: %v", err)
	}
	return value
}
