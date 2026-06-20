package main

import (
	"bytes"
	"fmt"
	"os"
)

func verifyDoc(repoRoot string, m manifest, want string) error {
	path := repoPath(repoRoot, m.GeneratedDoc)
	got, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if !bytes.Equal(got, []byte(want)) {
		return fmt.Errorf("%s is stale; run go run ./tools/integrationmatrix -write-doc", m.GeneratedDoc)
	}
	return nil
}
