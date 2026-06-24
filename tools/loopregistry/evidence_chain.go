package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type evidenceGraphDoc struct {
	Chains []evidenceGraphChain `json:"chains"`
}

type evidenceGraphChain struct {
	ID     string   `json:"id"`
	Claims []string `json:"claims,omitempty"`
}

func claimEvidenceChains(root string) (map[string][]string, error) {
	data, err := os.ReadFile(repoPath(root, evidenceGraphManifest))
	if err != nil {
		return nil, fmt.Errorf("read evidence graph manifest: %w", err)
	}
	var doc evidenceGraphDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("decode evidence graph manifest: %w", err)
	}
	chains := map[string][]string{}
	for _, chain := range doc.Chains {
		if chain.ID == "" {
			return nil, fmt.Errorf("evidence graph chain id is required")
		}
		for _, claimID := range chain.Claims {
			chains[claimID] = append(chains[claimID], chain.ID)
		}
	}
	return chains, nil
}

func verifyClaimEvidenceChains(claims []claimBinding, chains map[string][]string) error {
	for _, claim := range claims {
		if len(chains[claim.ID]) == 0 {
			return fmt.Errorf("claim %s has no evidence graph chain", claim.ID)
		}
	}
	return nil
}
