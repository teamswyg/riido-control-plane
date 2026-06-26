package main

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
	if err := maybeDoc(root, m, renderDoc(m, result), opt.WriteDoc, opt.CheckDoc); err != nil {
		return err
	}
	impact, err := maybeVerifyImpact(root, opt.Manifest, opt.ImpactBase, m)
	if err != nil {
		return err
	}
	if opt.GitHubAnnotations {
		writeGitHubAnnotations(opt.AnnotationOut, impact)
	}
	if opt.EvidenceOut != "" {
		return writeJSON(opt.EvidenceOut, newEvidence(m, result, impact))
	}
	return nil
}
