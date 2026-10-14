package main

func run(opt options) error {
	root, err := findRepoRoot(opt.Repo)
	if err != nil {
		return err
	}
	var m manifest
	if err := readJSON(repoPath(root, opt.Manifest), &m); err != nil {
		return err
	}
	if err := verifyAll(root, m); err != nil {
		return err
	}
	now, err := readinessNow()
	if err != nil {
		return err
	}
	e := newEvidenceAt(m, now)
	if err := maybeDoc(root, m, renderDoc(m, e), opt.WriteDoc, opt.CheckDoc); err != nil {
		return err
	}
	if opt.EvidenceOut != "" {
		if err := writeJSON(opt.EvidenceOut, e); err != nil {
			return err
		}
	}
	if opt.PublicStatusOut != "" {
		if err := writeText(opt.PublicStatusOut, renderPublicStatusDoc(e.PublicStatus)); err != nil {
			return err
		}
	}
	if opt.PublicStatusJSON != "" {
		if err := writeJSON(opt.PublicStatusJSON, e.PublicStatus); err != nil {
			return err
		}
	}
	if opt.PublicStatusHTML != "" {
		html, err := renderPublicStatusHTML(e.PublicStatus)
		if err != nil {
			return err
		}
		if err := writeText(opt.PublicStatusHTML, html); err != nil {
			return err
		}
	}
	if opt.PublicStatusBadgeJSON != "" {
		if err := writeJSON(opt.PublicStatusBadgeJSON, newPublicStatusBadge(e.PublicStatus)); err != nil {
			return err
		}
	}
	if opt.PublicStatusAnnotationOut != "" {
		body := renderPublicStatusGitHubAnnotation(e.PublicStatus)
		if err := writeText(opt.PublicStatusAnnotationOut, body); err != nil {
			return err
		}
	}
	if opt.CandidateOut != "" {
		return writeJSON(opt.CandidateOut, newCandidateEvidence(m, e, now))
	}
	return nil
}
