package main

import (
	"fmt"
	"os"
)

func run(opts options) error {
	root, err := findRepoRoot(opts.Repo)
	if err != nil {
		return err
	}
	m, err := loadManifest(resolve(root, opts.Manifest))
	if err != nil {
		return err
	}
	profile, err := evidenceProfileFor(m, opts.Profile)
	if err != nil {
		return err
	}
	if opts.WriteDoc {
		if err := writeText(resolve(root, m.GeneratedDoc), renderDoc(m)); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if err := verify(root, m, opts.CheckDoc); err != nil {
		return err
	}
	if opts.EvidenceOut != "" {
		if err := writeJSON(resolve(root, opts.EvidenceOut), newEvidence(m, profile)); err != nil {
			return err
		}
	}
	return nil
}

type options struct {
	Repo        string
	Manifest    string
	Profile     string
	EvidenceOut string
	WriteDoc    bool
	CheckDoc    bool
}

func mustRun(args []string) {
	if err := mainRun(args); err != nil {
		fmt.Fprintln(os.Stderr, "agentcatalogrbac:", err)
		os.Exit(1)
	}
}
