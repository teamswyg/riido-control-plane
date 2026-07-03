package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/reviewaccountseed/pathutil"
	"github.com/teamswyg/riido-control-plane/tools/reviewaccountseed/runconfig"
)

func run(opts runconfig.Options) error {
	root, err := pathutil.FindRepoRoot(opts.Repo)
	if err != nil {
		return err
	}
	m, err := loadManifest(pathutil.Resolve(root, opts.Manifest))
	if err != nil {
		return err
	}
	if err := verifySeedTerms(root, m); err != nil {
		return err
	}
	results, err := verifyCases(m.Cases)
	if err != nil {
		return err
	}
	if opts.WriteDoc {
		if err := writeText(pathutil.Resolve(root, m.GeneratedDoc), renderDoc(m)); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if err := verify(root, m, results, opts.CheckDoc); err != nil {
		return err
	}
	if opts.EvidenceOut != "" {
		return writeJSON(pathutil.Resolve(root, opts.EvidenceOut), newEvidence(m, results))
	}
	return nil
}
