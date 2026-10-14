package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/apiclientdelivery/doccheck"
	"github.com/teamswyg/riido-control-plane/tools/apiclientdelivery/pathutil"
	"github.com/teamswyg/riido-control-plane/tools/apiclientdelivery/runconfig"
)

func run(opt runconfig.Options) error {
	repoRoot, err := pathutil.FindRepoRoot(opt.Repo)
	if err != nil {
		return err
	}
	m, err := loadManifest(pathutil.Resolve(repoRoot, opt.Manifest))
	if err != nil {
		return err
	}
	result, err := verifyAll(repoRoot, m)
	if err != nil {
		return err
	}
	doc := renderDoc(m, result)
	if opt.WriteDoc {
		if err := writeText(pathutil.Resolve(repoRoot, m.GeneratedDoc), doc); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if opt.CheckDoc {
		if err := doccheck.Verify(pathutil.Resolve(repoRoot, m.GeneratedDoc), doc); err != nil {
			return err
		}
	}
	if opt.EvidenceOut != "" {
		return writeJSON(opt.EvidenceOut, newEvidence(m, result))
	}
	return nil
}
