package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/saascontrolplane/pathutil"
)

type options struct {
	Repo        string
	Manifest    string
	Boundary    string
	EvidenceOut string
	WriteDoc    bool
	CheckDoc    bool
}

func run(opt options) error {
	m, err := loadManifest(pathutil.RepoPath(opt.Repo, opt.Manifest))
	if err != nil {
		return err
	}
	if err := verifyManifest(opt.Repo, m); err != nil {
		return err
	}
	selected, err := selectBoundary(m, opt.Boundary)
	if err != nil {
		return err
	}
	doc := renderDoc(m)
	if err := verifyRequiredPhrases(opt.Repo, m.GeneratedDoc, doc, m.RequiredPhrases); err != nil {
		return err
	}
	if opt.WriteDoc {
		if err := writeText(pathutil.RepoPath(opt.Repo, m.GeneratedDoc), doc); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if opt.CheckDoc {
		if err := verifyDoc(opt.Repo, m, doc); err != nil {
			return err
		}
	}
	if opt.EvidenceOut != "" {
		if err := writeJSON(opt.EvidenceOut, newEvidence(m, selected)); err != nil {
			return fmt.Errorf("write evidence: %w", err)
		}
	}
	return nil
}
