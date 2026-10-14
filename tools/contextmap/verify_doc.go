package main

import (
	"fmt"
	"os"
)

func verifyDoc(root string, m manifest) error {
	path := resolve(root, m.GeneratedDoc)
	got, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if want := renderDoc(m); string(got) != want {
		return fmt.Errorf("generated doc is stale: run go run ./tools/contextmap -write-doc")
	}
	return nil
}
