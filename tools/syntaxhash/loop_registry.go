package main

import "fmt"

type loopRegistryClaim struct {
	ID           string `json:"id"`
	SemanticHash string `json:"semantic_hash"`
}

type loopRegistryManifest struct {
	Claims []loopRegistryClaim `json:"claim_bindings"`
}

func semanticHashes(root, path string) (map[string]string, error) {
	var m loopRegistryManifest
	if err := readJSON(repoPath(root, path), &m); err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, claim := range m.Claims {
		if claim.ID != "" && claim.SemanticHash != "" {
			out[claim.ID] = claim.SemanticHash
		}
	}
	return out, nil
}

func semanticHashFor(values map[string]string, claimID string) (string, error) {
	value := values[claimID]
	if value == "" {
		return "", fmt.Errorf("semantic claim %s has no hash", claimID)
	}
	return value, nil
}
