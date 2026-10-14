package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func readText(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

func writeText(path, value string) error {
	return os.WriteFile(path, []byte(value), 0o644)
}

func repoPath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
