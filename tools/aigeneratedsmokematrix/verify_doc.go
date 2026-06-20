package main

import (
	"fmt"
	"os"
)

func verifyDoc(repo string, m manifest, want string) error {
	got, err := os.ReadFile(repoPath(repo, m.GeneratedDoc))
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(got) != want {
		return fmt.Errorf("generated doc is stale: run go run ./tools/aigeneratedsmokematrix -write-doc")
	}
	return nil
}
