package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func repoPath(root, path string) string {
	return filepath.Join(root, filepath.FromSlash(path))
}

func readText(path string) (string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(body), nil
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal evidence: %w", err)
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}

func writeText(path, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}
	return os.WriteFile(path, []byte(value), 0o644)
}
