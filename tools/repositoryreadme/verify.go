package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/repositoryreadme/pathutil"
)

func verifyAll(root string, m manifest, rendered string) error {
	if err := verifyShape(m); err != nil {
		return err
	}
	if _, err := os.Stat(pathutil.RepoPath(root, m.Workflow)); err != nil {
		return fmt.Errorf("missing workflow %q: %w", m.Workflow, err)
	}
	for _, link := range m.DocLinks {
		if _, err := os.Stat(pathutil.RepoPath(root, link.Path)); err != nil {
			return fmt.Errorf("missing linked doc %q: %w", link.Path, err)
		}
	}
	return verifyRendered(rendered, m)
}

func verifyShape(m manifest) error {
	if m.SchemaVersion != manifestSchema || m.ID == "" || m.GeneratedDoc != generatedDoc {
		return fmt.Errorf("invalid manifest identity or generated_doc")
	}
	if m.Title == "" || m.RiidoTask == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		return fmt.Errorf("title, riido_task, workflow, and evidence_artifact are required")
	}
	if len(m.Summary) == 0 || len(m.DocLinks) == 0 || len(m.Verification) == 0 {
		return fmt.Errorf("summary, doc_links, and verification must not be empty")
	}
	if !completeLoop(m.Loop) {
		return fmt.Errorf("loop must define observe/hypothesis/execute/evaluate/retrospective")
	}
	return nil
}

func verifyRendered(rendered string, m manifest) error {
	for _, want := range m.RequiredMarkers {
		if !strings.Contains(rendered, want) {
			return fmt.Errorf("generated README missing required marker %q", want)
		}
	}
	for _, forbidden := range m.ForbiddenLiterals {
		if strings.Contains(rendered, forbidden) {
			return fmt.Errorf("generated README contains forbidden literal %q", forbidden)
		}
	}
	return nil
}
