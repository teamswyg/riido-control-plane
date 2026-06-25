package main

import "fmt"

type options struct {
	Repo        string
	CommandsIn  string
	DispatchOut string
}

func run(opt options) error {
	if opt.CommandsIn == "" {
		return fmt.Errorf("-commands-in is required")
	}
	if opt.DispatchOut == "" {
		return fmt.Errorf("-dispatch-out is required")
	}
	root, err := findRepoRoot(opt.Repo)
	if err != nil {
		return err
	}
	source, err := loadRefreshCommands(repoPath(root, opt.CommandsIn))
	if err != nil {
		return err
	}
	if source.SchemaVersion != refreshCommandsSchema {
		return fmt.Errorf("unsupported schema_version %q", source.SchemaVersion)
	}
	plan, err := buildDispatchPlan(root, source)
	if err != nil {
		return err
	}
	return writeJSON(repoPath(root, opt.DispatchOut), plan)
}
