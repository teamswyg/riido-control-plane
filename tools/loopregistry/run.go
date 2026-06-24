package main

import "fmt"

type options struct {
	Repo        string
	Manifest    string
	EvidenceOut string
	WriteDoc    bool
	CheckDoc    bool
	WriteHashes bool
}

func run(opt options) error {
	root, err := findRepoRoot(opt.Repo)
	if err != nil {
		return err
	}
	m, err := loadManifest(repoPath(root, opt.Manifest))
	if err != nil {
		return err
	}
	hashes, err := claimHashes(root, m)
	if err != nil {
		return err
	}
	if opt.WriteHashes {
		applyClaimHashes(&m, hashes)
		if err := writeJSON(repoPath(root, opt.Manifest), m); err != nil {
			return fmt.Errorf("write manifest hashes: %w", err)
		}
	}
	result, err := verifyAll(root, m, hashes)
	if err != nil {
		return err
	}
	doc := renderDoc(m, result)
	if err := maybeDoc(root, m.GeneratedDoc, doc, opt.WriteDoc, opt.CheckDoc); err != nil {
		return err
	}
	if opt.EvidenceOut != "" {
		return writeJSON(opt.EvidenceOut, newEvidence(m, result))
	}
	return nil
}
