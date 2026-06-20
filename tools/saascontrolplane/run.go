package main

import "fmt"

type options struct {
	Repo        string
	Manifest    string
	EvidenceOut string
	WriteDoc    bool
	CheckDoc    bool
}

func run(opt options) error {
	m, err := loadManifest(repoPath(opt.Repo, opt.Manifest))
	if err != nil {
		return err
	}
	if err := verifyManifest(opt.Repo, m); err != nil {
		return err
	}
	doc := renderDoc(m)
	if err := verifyRequiredPhrases(opt.Repo, m.GeneratedDoc, doc, m.RequiredPhrases); err != nil {
		return err
	}
	if opt.WriteDoc {
		if err := writeText(repoPath(opt.Repo, m.GeneratedDoc), doc); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if opt.CheckDoc {
		if err := verifyDoc(opt.Repo, m, doc); err != nil {
			return err
		}
	}
	if opt.EvidenceOut != "" {
		if err := writeJSON(opt.EvidenceOut, newEvidence(m)); err != nil {
			return fmt.Errorf("write evidence: %w", err)
		}
	}
	return nil
}
