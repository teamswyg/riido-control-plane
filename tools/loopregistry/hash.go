package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
)

func claimHashes(root string, m manifest) (map[string]string, error) {
	hashes := map[string]string{}
	for _, claim := range m.Claims {
		hash, err := claimHash(root, claim)
		if err != nil {
			return nil, err
		}
		hashes[claim.ID] = hash
	}
	return hashes, nil
}

func claimHash(root string, claim claimBinding) (string, error) {
	sum := sha256.New()
	writeHashPart(sum, "id", claim.ID)
	writeHashPart(sum, "statement", claim.Statement)
	for _, value := range sortedCopy(claim.CoversObserves) {
		writeHashPart(sum, "covers_observe", value)
	}
	for _, value := range sortedCopy(claim.CoversVerifies) {
		writeHashPart(sum, "covers_verify", value)
	}
	for _, value := range sortedCopy(claim.CoversFails) {
		writeHashPart(sum, "covers_fail", value)
	}
	for _, value := range sortedCopy(claim.CoversEvidence) {
		writeHashPart(sum, "covers_evidence", value)
	}
	for _, value := range sortedCopy(claim.Verifiers) {
		writeHashPart(sum, "verifier", value)
	}
	for _, value := range sortedCopy(claim.GeneratedDoc) {
		writeHashPart(sum, "doc", value)
	}
	for _, path := range sortedCopy(claim.Files) {
		data, err := os.ReadFile(repoPath(root, path))
		if err != nil {
			return "", fmt.Errorf("read claim file %s: %w", path, err)
		}
		writeHashPart(sum, "file:"+path, normalizedText(data))
	}
	return hex.EncodeToString(sum.Sum(nil)), nil
}

func sortedCopy(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
