package main

import (
	"fmt"
	"os"
)

func maybeDoc(root, path, doc string, writeDoc, checkDoc bool) error {
	fullPath := repoPath(root, path)
	if writeDoc {
		if err := os.WriteFile(fullPath, []byte(doc), 0o644); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if !checkDoc {
		return nil
	}
	current, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(current) != doc {
		return fmt.Errorf("generated doc drift: run go run ./tools/loopregistry -write-doc")
	}
	return nil
}
