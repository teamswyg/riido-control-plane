package main

import "fmt"

func verifyBuildStage(contract imageContract, parsed dockerfile) (int, *stage, error) {
	checks := 0
	require := countedRequire(&checks)
	if err := verifyBuildArg(contract, parsed, require); err != nil {
		return checks, nil, err
	}
	buildStage := parsed.stageByAlias(contract.Build.StageName)
	if buildStage == nil {
		return checks, nil, fmt.Errorf("build stage %q not found", contract.Build.StageName)
	}
	if err := verifyBuildStageBase(contract, buildStage, require); err != nil {
		return checks, nil, err
	}
	if err := verifyBuildModuleDownload(contract, buildStage, require); err != nil {
		return checks, nil, err
	}
	if err := verifyBuildCommand(contract, buildStage, require); err != nil {
		return checks, nil, err
	}
	return checks, buildStage, nil
}

func verifyBuildArg(contract imageContract, parsed dockerfile, require requireFunc) error {
	name := contract.Build.BuildArg.Name
	return require(parsed.Args[name] == contract.Build.BuildArg.Default,
		"ARG %s default = %q, want %q", name, parsed.Args[name], contract.Build.BuildArg.Default)
}

func verifyBuildStageBase(contract imageContract, buildStage *stage, require requireFunc) error {
	if err := require(buildStage.Base == "${"+contract.Build.BuildArg.Name+"}",
		"build stage base = %q, want ${%s}", buildStage.Base, contract.Build.BuildArg.Name); err != nil {
		return err
	}
	if err := require(buildStage.Workdir == contract.Build.Workdir,
		"build WORKDIR = %q, want %q", buildStage.Workdir, contract.Build.Workdir); err != nil {
		return err
	}
	return require(buildStage.Env["CGO_ENABLED"] == contract.Build.CGOEnabled,
		"CGO_ENABLED = %q, want %q", buildStage.Env["CGO_ENABLED"], contract.Build.CGOEnabled)
}
