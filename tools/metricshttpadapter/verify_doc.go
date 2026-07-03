package main

import (
	"fmt"
	"os"

	"github.com/teamswyg/riido-control-plane/tools/metricshttpadapter/pathutil"
)

func verifyDoc(root string, m manifest) error {
	current, err := os.ReadFile(pathutil.Resolve(root, m.GeneratedDoc))
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(current) != renderDoc(m) {
		return fmt.Errorf("%s is stale; run go run ./tools/metricshttpadapter -write-doc", m.GeneratedDoc)
	}
	return nil
}
