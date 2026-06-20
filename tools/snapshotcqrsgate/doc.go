package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func writeDocFile(repoRoot string, m manifest) error {
	path := filepath.Join(repoRoot, filepath.FromSlash(m.HumanDoc))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create doc dir: %w", err)
	}
	return os.WriteFile(path, []byte(renderDoc(m)), 0o644)
}
