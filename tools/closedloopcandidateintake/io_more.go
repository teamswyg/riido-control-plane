package main

import (
	"encoding/json"
	"os"
)

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func loadCandidate(path string) (candidateEvidence, []byte, error) {
	var candidate candidateEvidence
	data, err := os.ReadFile(path)
	if err != nil {
		return candidate, nil, err
	}
	return candidate, data, json.Unmarshal(data, &candidate)
}
