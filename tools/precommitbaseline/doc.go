package main

import (
	"bytes"
	"fmt"
	"os"
)

func verifyDoc(root string, m manifest, want string) error {
	got, err := os.ReadFile(repoPath(root, m.GeneratedDoc))
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if !bytes.Equal(got, []byte(want)) {
		return fmt.Errorf("%s is stale; run go run ./tools/precommitbaseline -write-doc", m.GeneratedDoc)
	}
	return nil
}
