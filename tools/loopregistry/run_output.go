package main

func writeRunOutputs(
	opt options,
	m manifest,
	result verifyResult,
	impact *impactEvidence,
) error {
	if !opt.GitHubAnnotations && !opt.TargetSummary &&
		opt.TargetScriptOut == "" && opt.EvidenceOut == "" {
		return nil
	}
	ev := newEvidence(m, result, impact)
	if opt.GitHubAnnotations {
		writeGitHubAnnotations(opt.AnnotationOut, result, ev.Impact)
	}
	if opt.TargetSummary {
		writeTargetVerifierSummary(opt.AnnotationOut, ev.Impact, opt.EvidenceOut)
	}
	if opt.TargetScriptOut != "" {
		if err := writeTargetVerifierScript(opt.TargetScriptOut, ev.Impact); err != nil {
			return err
		}
	}
	if opt.EvidenceOut != "" {
		return writeJSON(opt.EvidenceOut, ev)
	}
	return nil
}
