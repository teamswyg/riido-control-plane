package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func loadManifest(path string) (manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return manifest{}, fmt.Errorf("open manifest: %w", err)
	}
	defer file.Close()
	var m manifest
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&m); err != nil {
		return manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return manifest{}, fmt.Errorf("manifest must contain one JSON value: %w", err)
	}
	if err := loadFragments(filepath.Dir(path), &m); err != nil {
		return manifest{}, err
	}
	return m, nil
}

func loadFragment(path string) (manifestFragment, error) {
	file, err := os.Open(path)
	if err != nil {
		return manifestFragment{}, fmt.Errorf("open fragment: %w", err)
	}
	defer file.Close()
	var fragment manifestFragment
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fragment); err != nil {
		return manifestFragment{}, fmt.Errorf("decode fragment: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return manifestFragment{}, fmt.Errorf("fragment must contain one JSON value: %w", err)
	}
	if fragment.SchemaVersion != fragmentSchema || fragment.ID == "" {
		return manifestFragment{}, fmt.Errorf("invalid fragment identity in %s", path)
	}
	return fragment, nil
}

func writeText(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}
