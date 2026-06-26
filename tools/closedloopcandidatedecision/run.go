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
	if err := requireCandidateInput(opt); err != nil {
		return err
	}
	if opt.CandidateIn != "" {
		candidateResult, err := verifyCandidateDecisions(root, m, opt.CandidateIn)
		if err != nil {
			return err
		}
		result.CandidateCount = candidateResult.CandidateCount
		result.DecisionIDs = candidateResult.DecisionIDs
		result.DecisionArtifacts = candidateResult.DecisionArtifacts
		result.CandidateSourceRefs = candidateResult.CandidateSourceRefs
		result.ConsumedCandidateArtifacts = candidateResult.ConsumedCandidateArtifacts
	}
	if err := maybeDoc(root, m.GeneratedDoc, renderDoc(m, result), opt.WriteDoc, opt.CheckDoc); err != nil {
		return err
	}
	if opt.CommandsOut != "" {
		if err := writeJSON(opt.CommandsOut, newRefreshCommandEvidence(result)); err != nil {
			return err
		}
	}
	if opt.EvidenceOut != "" {
		return writeJSON(opt.EvidenceOut, newEvidence(m, result))
	}
	return nil
}

func requireCandidateInput(opt options) error {
	if opt.CandidateIn != "" || (!opt.CheckDoc && opt.EvidenceOut == "" && opt.CommandsOut == "") {
		return nil
	}
	return errMissingCandidateInput
}
