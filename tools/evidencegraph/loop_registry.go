package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type loopRegistryManifest struct {
	Loops []loopRegistryLoop `json:"loops"`
}

type loopRegistryLoop struct {
	ID string `json:"id"`
}

func loadLoopRegistryIDs(root, path string) (map[string]bool, error) {
	data, err := os.ReadFile(repoPath(root, path))
	if err != nil {
		return nil, fmt.Errorf("read loop registry manifest: %w", err)
	}
	var m loopRegistryManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("decode loop registry manifest: %w", err)
	}
	out := map[string]bool{}
	for _, loop := range m.Loops {
		if loop.ID != "" {
			out[loop.ID] = true
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("loop registry manifest has no loops")
	}
	return out, nil
}
