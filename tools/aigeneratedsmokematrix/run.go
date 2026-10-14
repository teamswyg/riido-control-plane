package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/aigeneratedsmokematrix/doccheck"
	"github.com/teamswyg/riido-control-plane/tools/aigeneratedsmokematrix/pathutil"
)

type options struct {
	Repo        string
	Manifest    string
	EvidenceOut string
	WriteDoc    bool
	CheckDoc    bool
}

func run(opt options) error {
	m, err := loadManifest(pathutil.Resolve(opt.Repo, opt.Manifest))
	if err != nil {
		return err
	}
	result, err := verifyAll(opt.Repo, m)
	if err != nil {
		return err
	}
	doc := renderDoc(m, result.Counts)
	if opt.WriteDoc {
		if err := writeText(pathutil.Resolve(opt.Repo, m.GeneratedDoc), doc); err != nil {
			return fmt.Errorf("write generated doc: %w", err)
		}
	}
	if opt.CheckDoc {
		if err := doccheck.Verify(pathutil.Resolve(opt.Repo, m.GeneratedDoc), doc); err != nil {
			return err
		}
	}
	if opt.EvidenceOut != "" {
		return writeJSON(opt.EvidenceOut, newEvidence(m, result.Counts))
	}
	return nil
}
