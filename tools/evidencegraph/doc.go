package main

import (
	"fmt"
	"os"
)

func maybeDoc(root string, m manifest, doc string, writeDoc, checkDoc bool) error {
	full := repoPath(root, m.GeneratedDoc)
	if writeDoc {
		if err := writeText(full, doc); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if !checkDoc {
		return nil
	}
	current, err := os.ReadFile(full)
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(current) != doc {
		return fmt.Errorf("generated doc drift: run go run ./tools/evidencegraph -write-doc")
	}
	return nil
}
