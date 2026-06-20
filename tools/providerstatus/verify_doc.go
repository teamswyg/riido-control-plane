package main

import (
	"fmt"
	"os"
)

func verifyDoc(root string, m manifest) error {
	got, err := os.ReadFile(resolve(root, m.GeneratedDoc))
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if want := renderDoc(m); string(got) != want {
		return fmt.Errorf("generated doc is stale: run go run ./tools/providerstatus -write-doc")
	}
	return nil
}
