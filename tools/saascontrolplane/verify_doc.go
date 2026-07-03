package main

import (
	"fmt"
	"os"

	"github.com/teamswyg/riido-control-plane/tools/saascontrolplane/pathutil"
)

func verifyDoc(repo string, m manifest, want string) error {
	got, err := os.ReadFile(pathutil.RepoPath(repo, m.GeneratedDoc))
	if err != nil {
		return fmt.Errorf("read generated doc: %w", err)
	}
	if string(got) != want {
		return fmt.Errorf("generated doc is stale: run go run ./tools/saascontrolplane -write-doc")
	}
	return nil
}

func verifyLoop(loop evidenceLoop) error {
	if loop.Observation == "" || loop.Hypothesis == "" || loop.Execute == "" ||
		loop.Evaluate == "" || loop.Retrospective == "" {
		return fmt.Errorf("evidence loop must define observe/hypothesis/execute/evaluate/retrospective")
	}
	return nil
}
