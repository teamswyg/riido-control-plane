package main

import (
	"encoding/json"
	"os"
	"strings"
)

type sourceManifestMeta struct {
	ID         string       `json:"id"`
	Assertions []string     `json:"assertions"`
	Loop       evidenceLoop `json:"loop"`
}

func sourceManifestHasMetadata(root, path string) bool {
	data, err := os.ReadFile(resolvePath(root, path))
	if err != nil {
		return false
	}
	var meta sourceManifestMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return false
	}
	return strings.TrimSpace(meta.ID) != "" &&
		hasNonEmptyAssertions(meta.Assertions) &&
		completeLoop(meta.Loop)
}

func hasNonEmptyAssertions(assertions []string) bool {
	if len(assertions) == 0 {
		return false
	}
	for _, assertion := range assertions {
		if strings.TrimSpace(assertion) == "" {
			return false
		}
	}
	return true
}
