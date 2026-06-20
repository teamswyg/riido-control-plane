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
		return fmt.Errorf("%s is stale; run go run ./tools/reviewaccountseed -write-doc", m.GeneratedDoc)
	}
	return nil
}
