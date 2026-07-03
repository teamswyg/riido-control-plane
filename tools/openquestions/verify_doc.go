package main

import (
	"fmt"
	"os"

	"github.com/teamswyg/riido-control-plane/tools/openquestions/pathutil"
)

func verifyDoc(repoRoot string, m manifest, want string) error {
	got, err := os.ReadFile(pathutil.RepoPath(repoRoot, m.GeneratedDoc))
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(got) != want {
		return fmt.Errorf("%s is stale; run go run ./tools/openquestions -write-doc", m.GeneratedDoc)
	}
	return nil
}
