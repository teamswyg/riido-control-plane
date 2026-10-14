package main

import (
	"encoding/json"
	"os"
)

type smokeMatrix struct {
	SchemaVersion string       `json:"schema_version"`
	Entries       []smokeEntry `json:"entries"`
}

type smokeEntry struct {
	GeneratedPath string   `json:"generated_path"`
	Method        string   `json:"method"`
	Path          string   `json:"path"`
	EvidenceTests []string `json:"evidence_tests"`
}

func loadSmokeMatrix(path string) (smokeMatrix, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return smokeMatrix{}, err
	}
	var matrix smokeMatrix
	if err := json.Unmarshal(body, &matrix); err != nil {
		return smokeMatrix{}, err
	}
	return matrix, nil
}
