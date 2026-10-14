package main

import (
	"fmt"
	"os"

	"github.com/teamswyg/riido-control-plane/tools/agentruntimebinding/pathutil"
)

func run(opts options) error {
	root, err := pathutil.FindRepoRoot(opts.Repo)
	if err != nil {
		return err
	}
	m, err := loadManifest(pathutil.Resolve(root, opts.Manifest))
	if err != nil {
		return err
	}
	if opts.WriteDoc {
		if err := writeText(pathutil.Resolve(root, m.GeneratedDoc), renderDoc(m)); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if err := verify(root, m, opts.CheckDoc); err != nil {
		return err
	}
	if opts.EvidenceOut != "" {
		if err := writeJSON(pathutil.Resolve(root, opts.EvidenceOut), newEvidence(m)); err != nil {
			return err
		}
	}
	return nil
}

type options struct {
	Repo        string
	Manifest    string
	EvidenceOut string
	WriteDoc    bool
	CheckDoc    bool
}

func mustRun(args []string) {
	if err := mainRun(args); err != nil {
		fmt.Fprintln(os.Stderr, "agentruntimebinding:", err)
		os.Exit(1)
	}
}
