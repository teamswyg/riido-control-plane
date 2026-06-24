package main

import (
	"encoding/json"
	"os"
)

func loadManifest(path string) (manifest, error) {
	var m manifest
	err := readJSON(path, &m)
	return m, err
}

func loadCandidate(path string) (candidateEvidence, []byte, error) {
	var candidate candidateEvidence
	data, err := os.ReadFile(path)
	if err != nil {
		return candidate, nil, err
	}
	return candidate, data, json.Unmarshal(data, &candidate)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}
