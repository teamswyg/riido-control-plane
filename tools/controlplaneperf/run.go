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
	if opt.AppendBenchmarkLog != "" {
		if err := appendBenchmarkHistory(root, m, opt.BenchmarkIn, opt.AppendBenchmarkLog); err != nil {
			return err
		}
	}
	if err := verifyAll(root, m); err != nil {
		return err
	}
	e := newEvidence(m)
	if err := maybeDoc(root, m, renderDoc(m, e), opt.WriteDoc, opt.CheckDoc); err != nil {
		return err
	}
	if opt.EvidenceOut != "" {
		if err := writeJSON(opt.EvidenceOut, e); err != nil {
			return err
		}
	}
	if opt.ArchitectureQueryOut != "" {
		if len(opt.ArchitecturePaths) == 0 {
			return errArchitectureQueryPathRequired()
		}
		query := newArchitectureQuery(m, opt.ArchitecturePaths)
		if err := writeJSON(opt.ArchitectureQueryOut, query); err != nil {
			return err
		}
	}
	return nil
}
