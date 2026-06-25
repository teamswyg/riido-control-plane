package main

func run(opt options) error {
	root, err := findRepoRoot(opt.Repo)
	if err != nil {
		return err
	}
	m, deps, err := loadAll(root, opt.Manifest)
	if err != nil {
		return err
	}
	if err := verifyAll(root, m, deps); err != nil {
		return err
	}
	e := newEvidence(m)
	if err := maybeDoc(root, m, renderDoc(m, e), opt.WriteDoc, opt.CheckDoc); err != nil {
		return err
	}
	if opt.EvidenceOut != "" {
		return writeJSON(opt.EvidenceOut, e)
	}
	return nil
}

func loadAll(root, manifestPath string) (manifest, dependencies, error) {
	var m manifest
	if err := readJSON(repoPath(root, manifestPath), &m); err != nil {
		return manifest{}, dependencies{}, err
	}
	deps := dependencies{}
	if err := readJSON(repoPath(root, m.LoopRegistryManifest), &deps.registry); err != nil {
		return manifest{}, dependencies{}, err
	}
	if err := readJSON(repoPath(root, m.EvidenceGraphManifest), &deps.graph); err != nil {
		return manifest{}, dependencies{}, err
	}
	if err := readJSON(repoPath(root, m.PreCommitManifest), &deps.preCommit); err != nil {
		return manifest{}, dependencies{}, err
	}
	return m, deps, nil
}

type dependencies struct {
	registry  loopRegistry
	graph     evidenceGraph
	preCommit preCommitManifest
}
