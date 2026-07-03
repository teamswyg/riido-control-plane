package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/metricshttpadapter/pathutil"
	"github.com/teamswyg/riido-control-plane/tools/metricshttpadapter/runconfig"
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
	result, err := exerciseAdapter(m)
	if err != nil {
		return err
	}
	if opts.WriteDoc {
		if err := writeText(pathutil.Resolve(root, m.GeneratedDoc), renderDoc(m)); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if err := verify(root, m, result, opts.CheckDoc); err != nil {
		return err
	}
	if opts.EvidenceOut != "" {
		return writeJSON(pathutil.Resolve(root, opts.EvidenceOut), newEvidence(m, result))
	}
	return nil
}
