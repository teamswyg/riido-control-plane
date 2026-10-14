package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func findRepoRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("abs repo: %w", err)
	}
	for {
		if pathExists(filepath.Join(dir, "go.mod")) || pathExists(filepath.Join(dir, ".git")) {
			return dir, nil
		}
		next := filepath.Dir(dir)
		if next == dir {
			return "", fmt.Errorf("repo root not found from %s", start)
		}
		dir = next
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func contains(items []string, needle string) bool {
	for _, item := range items {
		if item == needle {
			return true
		}
	}
	return false
}

func containsText(items []string, needle string) bool {
	for _, item := range items {
		if strings.Contains(item, needle) {
			return true
		}
	}
	return false
}

func nonEmpty(values ...string) bool {
	for _, value := range values {
		if value == "" {
			return false
		}
	}
	return true
}
