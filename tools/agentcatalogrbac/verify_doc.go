package main

import (
	"fmt"
	"os"
)

func verifyDoc(root string, m manifest) error {
	current, err := os.ReadFile(resolve(root, m.GeneratedDoc))
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(current) != renderDoc(m) {
		return fmt.Errorf("generated doc drift: run go run ./tools/agentcatalogrbac -write-doc")
	}
	return nil
}
