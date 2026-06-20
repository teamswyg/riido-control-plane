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
	result, err := verifyAll(opt.Repo, m)
	if err != nil {
		return err
	}
	doc := renderDoc(m, result.Counts)
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
		return writeJSON(opt.EvidenceOut, newEvidence(m, result.Counts))
	}
	return nil
}
