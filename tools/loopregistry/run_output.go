package main

func writeRunOutputs(
	opt options,
	m manifest,
	result verifyResult,
	impact *impactEvidence,
) error {
	if !opt.GitHubAnnotations && !opt.TargetSummary && opt.EvidenceOut == "" {
		return nil
	}
	ev := newEvidence(m, result, impact)
	if opt.GitHubAnnotations {
		writeGitHubAnnotations(opt.AnnotationOut, result, ev.Impact)
	}
	if opt.TargetSummary {
		writeTargetVerifierSummary(opt.AnnotationOut, ev.Impact, opt.EvidenceOut)
	}
	if opt.EvidenceOut != "" {
		return writeJSON(opt.EvidenceOut, ev)
	}
	return nil
}
