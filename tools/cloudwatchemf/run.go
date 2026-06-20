package main

import "fmt"

func run(opts options) error {
	root, err := findRepoRoot(opts.Repo)
	if err != nil {
		return err
	}
	m, err := loadManifest(resolve(root, opts.Manifest))
	if err != nil {
		return err
	}
	shape, err := buildEMFShape()
	if err != nil {
		return err
	}
	if opts.WriteDoc {
		if err := writeText(resolve(root, m.GeneratedDoc), renderDoc(m)); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if err := verify(root, m, shape, opts.CheckDoc); err != nil {
		return err
	}
	if opts.EvidenceOut != "" {
		if err := writeJSON(resolve(root, opts.EvidenceOut), newEvidence(m, shape)); err != nil {
			return err
		}
	}
	return nil
}
