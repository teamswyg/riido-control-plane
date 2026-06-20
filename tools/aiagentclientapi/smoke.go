package main

import (
	"encoding/json"
	"os"
	"strings"
)

type smokeMatrix struct {
	Entries []smokeEntry `json:"entries"`
}

type smokeEntry struct {
	GeneratedPath string `json:"generated_path"`
}

func loadSmokeGeneratedPaths(path string) (map[string]struct{}, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	var matrix smokeMatrix
	if err := json.Unmarshal(data, &matrix); err != nil {
		return nil, 0, err
	}
	paths := map[string]struct{}{}
	for _, entry := range matrix.Entries {
		paths[strings.TrimSpace(entry.GeneratedPath)] = struct{}{}
	}
	return paths, len(matrix.Entries), nil
}
