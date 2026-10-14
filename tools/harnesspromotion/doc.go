package main

import (
	"fmt"
	"os"
)

func maybeDoc(root, path, doc string, writeDoc, checkDoc bool) error {
	full := repoPath(root, path)
	if writeDoc {
		if err := os.WriteFile(full, []byte(doc), 0o644); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if !checkDoc {
		return nil
	}
	current, err := os.ReadFile(full)
	if err != nil {
		return err
	}
	if string(current) != doc {
		return fmt.Errorf("generated doc drift: run go run ./tools/harnesspromotion -write-doc")
	}
	return nil
}
