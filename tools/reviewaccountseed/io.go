package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func loadManifest(path string) (manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest{}, err
	}
	var out manifest
	if err := json.Unmarshal(data, &out); err != nil {
		return manifest{}, err
	}
	return out, nil
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}

func writeText(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}

func mustRun(args []string) {
	if err := mainRun(args); err != nil {
		fmt.Fprintln(os.Stderr, "reviewaccountseed:", err)
		os.Exit(1)
	}
}
