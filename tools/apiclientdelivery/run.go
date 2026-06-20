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
	repoRoot, err := findRepoRoot(opt.Repo)
	if err != nil {
		return err
	}
	m, err := loadManifest(repoPath(repoRoot, opt.Manifest))
	if err != nil {
		return err
	}
	result, err := verifyAll(repoRoot, m)
	if err != nil {
		return err
	}
	doc := renderDoc(m, result)
	if opt.WriteDoc {
		if err := writeText(repoPath(repoRoot, m.GeneratedDoc), doc); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if opt.CheckDoc {
		if err := verifyDoc(repoRoot, m, doc); err != nil {
			return err
		}
	}
	if opt.EvidenceOut != "" {
		return writeJSON(opt.EvidenceOut, newEvidence(m, result))
	}
	return nil
}
