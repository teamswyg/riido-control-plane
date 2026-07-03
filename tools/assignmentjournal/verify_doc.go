package main

import (
	"fmt"
	"os"

	"github.com/teamswyg/riido-control-plane/tools/assignmentjournal/pathutil"
)

func verifyDoc(root string, m manifest) error {
	current, err := os.ReadFile(pathutil.Resolve(root, m.GeneratedDoc))
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(current) != renderDoc(m) {
		return fmt.Errorf("generated doc drift: run go run ./tools/assignmentjournal -write-doc")
	}
	return nil
}
