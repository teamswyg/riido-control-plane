package main

import (
	"os"
	"path/filepath"
)

func findRepoRoot(start string) (string, error) {
	path, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
			return path, nil
		}
		next := filepath.Dir(path)
		if next == path {
			return "", os.ErrNotExist
		}
		path = next
	}
}

func repoPath(root, rel string) string {
	return filepath.Join(root, filepath.FromSlash(rel))
}
