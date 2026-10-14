package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-control-plane/tools/repositoryreadme/pathutil"
)

func runWithOptions(repoRoot, manifestPath string, writeDoc, checkDoc bool, evidenceOut string) error {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}
	m, err := loadManifest(pathutil.RepoPath(root, manifestPath))
	if err != nil {
		return err
	}
	rendered := renderDoc(m)
	if err := verifyAll(root, m, rendered); err != nil {
		return err
	}
	if writeDoc {
		if err := writeText(pathutil.RepoPath(root, generatedDoc), rendered); err != nil {
			return err
		}
	}
	if checkDoc {
		current, err := os.ReadFile(pathutil.RepoPath(root, generatedDoc))
		if err != nil {
			return err
		}
		if string(current) != rendered {
			return fmt.Errorf("%s is stale; run go run ./tools/repositoryreadme -write-doc", generatedDoc)
		}
	}
	if evidenceOut != "" {
		return writeJSON(pathutil.RepoPath(root, evidenceOut), buildEvidence(m))
	}
	return nil
}
