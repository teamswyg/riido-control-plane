package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyRefs(root, chainID, group string, refs []ref) error {
	for _, item := range refs {
		if err := verifyRef(root, chainID, group, item); err != nil {
			return err
		}
	}
	return nil
}

func verifyRef(root, chainID, group string, item ref) error {
	if item.Kind == "" || item.Path == "" {
		return fmt.Errorf("%s %s ref kind and path are required", chainID, group)
	}
	if !knownEvidenceRefKind(item.Kind) {
		return fmt.Errorf("%s %s ref %s has unsupported kind %q", chainID, group, item.Path, item.Kind)
	}
	if item.Kind == "artifact" {
		return verifyArtifactRef(chainID, item)
	}
	if item.Redacted {
		return fmt.Errorf("%s %s ref %s is redacted but is not an artifact", chainID, group, item.Path)
	}
	if _, err := os.Stat(repoPath(root, item.Path)); err != nil {
		return fmt.Errorf("%s %s ref %s missing: %w", chainID, group, item.Path, err)
	}
	return nil
}

func verifyArtifactRef(chainID string, item ref) error {
	if !item.Redacted {
		return fmt.Errorf("%s artifact ref %s must be redacted", chainID, item.Path)
	}
	if strings.Contains(item.Path, "/") || strings.Contains(item.Path, "://") {
		return fmt.Errorf("%s artifact ref %s must be an artifact name only", chainID, item.Path)
	}
	return nil
}
