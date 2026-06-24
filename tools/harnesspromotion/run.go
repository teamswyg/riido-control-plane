package main

import "fmt"

func run(opt options) error {
	root, err := findRepoRoot(opt.Repo)
	if err != nil {
		return err
	}
	m, err := loadManifest(repoPath(root, opt.Manifest))
	if err != nil {
		return err
	}
	result, err := verifyAll(root, m)
	if err != nil {
		return err
	}
	if err := maybeDoc(root, m.GeneratedDoc, renderDoc(m, result), opt.WriteDoc, opt.CheckDoc); err != nil {
		return err
	}
	if opt.Summary != "" {
		if opt.CandidateOut == "" {
			return fmt.Errorf("candidate-out is required with summary")
		}
		if err := promoteSummary(root, m, opt.Summary, opt.CandidateOut); err != nil {
			return err
		}
	}
	if opt.EvidenceOut != "" {
		return writeJSON(opt.EvidenceOut, newEvidence(m, result))
	}
	return nil
}
