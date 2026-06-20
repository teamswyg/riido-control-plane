package main

import "fmt"

func verifyDoc(root, expected string) error {
	actual, err := readText(repoPath(root, generatedDoc))
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("%s is stale; run go run ./tools/runtimecdownership -write-doc", generatedDoc)
	}
	return nil
}
