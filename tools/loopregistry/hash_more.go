package main

import (
	"hash"
	"io"
)

func writeHashPart(sum hash.Hash, key, value string) {
	_, _ = io.WriteString(sum, key)
	_, _ = io.WriteString(sum, "\x00")
	_, _ = io.WriteString(sum, value)
	_, _ = io.WriteString(sum, "\x00")
}

func applyClaimHashes(m *manifest, hashes map[string]string) {
	for i := range m.Claims {
		if hash := hashes[m.Claims[i].ID]; hash != "" {
			m.Claims[i].SemanticHash = hash
		}
	}
}
