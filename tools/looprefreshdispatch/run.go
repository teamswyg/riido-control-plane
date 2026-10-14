package main

import "fmt"

type options struct {
	Repo         string
	CommandsIn   commandInputPaths
	DispatchOut  string
	CandidateOut string
}

func run(opt options) error {
	if len(opt.CommandsIn) == 0 {
		return fmt.Errorf("-commands-in is required")
	}
	if opt.DispatchOut == "" {
		return fmt.Errorf("-dispatch-out is required")
	}
	root, err := findRepoRoot(opt.Repo)
	if err != nil {
		return err
	}
	sources, err := loadRefreshCommandSources(root, opt.CommandsIn)
	if err != nil {
		return err
	}
	source, err := mergeRefreshCommandSources(sources)
	if err != nil {
		return err
	}
	plan, err := buildDispatchPlan(root, source)
	if err != nil {
		return err
	}
	if err := writeJSON(repoPath(root, opt.DispatchOut), plan); err != nil {
		return err
	}
	if opt.CandidateOut != "" {
		return writeJSON(repoPath(root, opt.CandidateOut), candidateEvidenceFromPlan(plan))
	}
	return nil
}
