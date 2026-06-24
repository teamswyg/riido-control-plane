package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func findRepoRoot(start string) (string, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(abs, ".git")); err == nil {
			return abs, nil
		}
		next := filepath.Dir(abs)
		if next == abs {
			return "", fmt.Errorf("repository root not found from %s", start)
		}
		abs = next
	}
}

func repoPath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.FromSlash(path))
}

func loadManifest(path string) (manifest, error) {
	var m manifest
	data, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	return m, json.Unmarshal(data, &m)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}
