package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest{}, err
	}
	return decodeManifest(strings.NewReader(string(data)))
}

func loadLoopRegistryEvidence(path string) (evidence, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return evidence{}, err
	}
	var got evidence
	if err := json.Unmarshal(data, &got); err != nil {
		return evidence{}, err
	}
	return got, nil
}

func decodeManifest(r io.Reader) (manifest, error) {
	var m manifest
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		return m, err
	}
	return m, nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func normalizedText(data []byte) string {
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}
