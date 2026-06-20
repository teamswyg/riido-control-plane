package main

import "fmt"

func verifyDoc(root string, m manifest, want string) error {
	got, err := readText(repoPath(root, m.GeneratedDoc))
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("%s is stale; run go run ./tools/liveworkflowevidence -write-doc", m.GeneratedDoc)
	}
	return nil
}
