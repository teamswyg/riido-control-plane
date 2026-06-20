package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func findRepoRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("abs repo: %w", err)
	}
	for {
		if exists(filepath.Join(dir, "go.mod")) || exists(filepath.Join(dir, ".git")) {
			return dir, nil
		}
		next := filepath.Dir(dir)
		if next == dir {
			return "", fmt.Errorf("repo root not found from %s", start)
		}
		dir = next
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
