package main

import (
	"fmt"
	"io"
)

type options struct {
	Repo        string
	Manifest    string
	EvidenceOut string
	WriteDoc    bool
	CheckDoc    bool
	WriteHashes bool
	ImpactBase  string

	GitHubAnnotations  bool
	TargetSummary      bool
	AnnotationOut      io.Writer
	RefreshPlanIn      string
	RefreshCommandsOut string
}

func run(opt options) error {
	if opt.RefreshPlanIn != "" {
		return writeRefreshCommandEvidence(opt)
	}
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
	impact, err := maybeVerifyImpact(root, opt.Manifest, opt.ImpactBase, m)
	if err != nil {
		return err
	}
	return writeRunOutputs(opt, m, result, impact)
}
