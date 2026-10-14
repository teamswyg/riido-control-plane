package main

import (
	"fmt"
	"os"

	"github.com/teamswyg/riido-control-plane/tools/providerstatus/pathutil"
)

func verifyDoc(root string, m manifest) error {
	got, err := os.ReadFile(pathutil.Resolve(root, m.GeneratedDoc))
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if want := renderDoc(m); string(got) != want {
		return fmt.Errorf("generated doc is stale: run go run ./tools/providerstatus -write-doc")
	}
	return nil
}
