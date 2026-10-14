package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

func loadManifest(path string) (evidenceManifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return evidenceManifest{}, fmt.Errorf("open manifest: %w", err)
	}
	defer file.Close()
	var manifest evidenceManifest
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return evidenceManifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return evidenceManifest{}, fmt.Errorf("manifest must contain a single JSON object: %w", err)
	}
	return manifest, nil
}
