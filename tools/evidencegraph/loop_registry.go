package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type loopRegistryManifest struct {
	Loops  []loopRegistryLoop  `json:"loops"`
	Claims []loopRegistryClaim `json:"claim_bindings"`
}

type loopRegistryLoop struct {
	ID string `json:"id"`
}

type loopRegistryClaim struct {
	ID string `json:"id"`
}

type loopRegistryIndex struct {
	Loops  map[string]bool
	Claims map[string]bool
}

func loadLoopRegistryIndex(root, path string) (loopRegistryIndex, error) {
	data, err := os.ReadFile(repoPath(root, path))
	if err != nil {
		return loopRegistryIndex{}, fmt.Errorf("read loop registry manifest: %w", err)
	}
	var m loopRegistryManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return loopRegistryIndex{}, fmt.Errorf("decode loop registry manifest: %w", err)
	}
	index := loopRegistryIndex{Loops: map[string]bool{}, Claims: map[string]bool{}}
	for _, loop := range m.Loops {
		if loop.ID != "" {
			index.Loops[loop.ID] = true
		}
	}
	for _, claim := range m.Claims {
		if claim.ID != "" {
			index.Claims[claim.ID] = true
		}
	}
	if len(index.Loops) == 0 || len(index.Claims) == 0 {
		return loopRegistryIndex{}, fmt.Errorf("loop registry manifest must have loops and claims")
	}
	return index, nil
}
