package main

import (
	"fmt"
	"os"
)

func verifyDoc(repoRoot string, m manifest, want string) error {
	got, err := os.ReadFile(repoPath(repoRoot, m.GeneratedDoc))
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(got) != want {
		return fmt.Errorf("%s is stale; run go run ./tools/apiclientdelivery -write-doc", m.GeneratedDoc)
	}
	return nil
}
