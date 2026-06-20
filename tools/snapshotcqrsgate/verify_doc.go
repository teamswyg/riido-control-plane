package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func verifyDoc(repoRoot string, m manifest) error {
	path := filepath.Join(repoRoot, filepath.FromSlash(m.HumanDoc))
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read human doc: %w", err)
	}
	if string(data) != renderDoc(m) {
		return fmt.Errorf("reader doc drift: run go run ./tools/snapshotcqrsgate -write-doc")
	}
	return nil
}
