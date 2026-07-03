package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/cloudwatchemf/pathutil"
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
	shape, err := buildEMFShape()
	if err != nil {
		return err
	}
	if opts.WriteDoc {
		if err := writeText(pathutil.Resolve(root, m.GeneratedDoc), renderDoc(m)); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if err := verify(root, m, shape, opts.CheckDoc); err != nil {
		return err
	}
	if opts.EvidenceOut != "" {
		if err := writeJSON(pathutil.Resolve(root, opts.EvidenceOut), newEvidence(m, shape)); err != nil {
			return err
		}
	}
	return nil
}
