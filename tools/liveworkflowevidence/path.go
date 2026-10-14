package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const defaultManifest = "docs/30-architecture/live-workflow-redaction.riido.json"

func findRepoRoot(start string) (string, error) {
	if start == "" {
		start = "."
	}
	root, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			return root, nil
		}
		next := filepath.Dir(root)
		if next == root {
			return "", fmt.Errorf("go.mod not found from %s", start)
		}
		root = next
	}
}

func repoPath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.FromSlash(path))
}
