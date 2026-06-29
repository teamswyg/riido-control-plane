package main

import "fmt"

func maybeDoc(root string, m manifest, doc string, writeDoc, checkDoc bool) error {
	path := repoPath(root, m.GeneratedDoc)
	if writeDoc {
		if err := writeText(path, doc); err != nil {
			return err
		}
	}
	if checkDoc {
		current, err := readText(path)
		if err != nil {
			return err
		}
		if current != doc {
			return fmt.Errorf("generated doc drift: run go run ./tools/operationalreadiness -write-doc")
		}
	}
	return nil
}
