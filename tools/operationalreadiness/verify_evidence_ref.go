package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyEvidenceRef(root string, ref evidenceRef) error {
	if ref.Kind == "" || ref.Path == "" {
		return fmt.Errorf("evidence ref must bind kind and path")
	}
	if strings.Contains(ref.Path, ":") {
		return nil
	}
	return requireLocalFile(root, ref.Path)
}

func requireLocalFile(root, path string) error {
	if _, err := os.Stat(repoPath(root, path)); err != nil {
		return fmt.Errorf("missing local evidence file %s: %w", path, err)
	}
	return nil
}
