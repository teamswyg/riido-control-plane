package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

func verifyDoc(repoRoot string, m manifest, want string) error {
	path := filepath.Join(repoRoot, filepath.FromSlash(m.GeneratedDoc))
	got, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if !bytes.Equal(got, []byte(want)) {
		return fmt.Errorf("%s is stale; run go run ./tools/migrationledger -write-doc", m.GeneratedDoc)
	}
	return nil
}
