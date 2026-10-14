package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/containercontract/dockerfile"
)

func verifyBuildStage(contract imageContract, parsed dockerfile.File) (int, *dockerfile.Stage, error) {
	checks := 0
	require := countedRequire(&checks)
	if err := verifyBuildArg(contract, parsed, require); err != nil {
		return checks, nil, err
	}
	buildStage := parsed.StageByAlias(contract.Build.StageName)
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

func verifyBuildArg(contract imageContract, parsed dockerfile.File, require requireFunc) error {
	name := contract.Build.BuildArg.Name
	return require(parsed.Args[name] == contract.Build.BuildArg.Default,
		"ARG %s default = %q, want %q", name, parsed.Args[name], contract.Build.BuildArg.Default)
}

func verifyBuildStageBase(contract imageContract, buildStage *dockerfile.Stage, require requireFunc) error {
	if err := require(buildStage.Base == "${"+contract.Build.BuildArg.Name+"}",
		"build stage base = %q, want ${%s}", buildStage.Base, contract.Build.BuildArg.Name); err != nil {
		return err
	}
	if err := require(buildStage.Workdir == contract.Build.Workdir,
		"build WORKDIR = %q, want %q", buildStage.Workdir, contract.Build.Workdir); err != nil {
		return err
	}
	return require(buildStage.Env["CGO_ENABLED"] == contract.Build.CGOEnabled, "CGO_ENABLED = %q, want %q", buildStage.Env["CGO_ENABLED"], contract.Build.CGOEnabled)
}

func verifyBuildModuleDownload(contract imageContract, buildStage *dockerfile.Stage, require requireFunc) error {
	moduleDownload := contract.Build.ModuleDownload
	if err := require(dockerfile.HasModuleDownloadRun(buildStage.Runs, moduleDownload.Command),
		"module download command %q not found", moduleDownload.Command); err != nil {
		return err
	}
	return requireCacheMounts(buildStage, "module download", moduleDownload.Command, moduleDownload.CacheMounts, require)
}

func verifyBuildCommand(contract imageContract, stage *dockerfile.Stage, require requireFunc) error {
	want := contract.Build.GoBuild
	if err := require(dockerfile.HasGoBuildRun(stage.Runs, want.Output, want.Package, want.Trimpath, want.LDFlags),
		"go build command does not satisfy container contract"); err != nil {
		return err
	}
	return requireCacheMounts(stage, "go build", "go build", contract.Build.GoBuild.CacheMounts, require)
}

func requireCacheMounts(buildStage *dockerfile.Stage, label, command string, targets []string, require requireFunc) error {
	for _, target := range targets {
		if err := require(dockerfile.RunHasCacheMount(buildStage.Runs, command, target),
			"%s cache mount target %q not found", label, target); err != nil {
			return err
		}
	}
	return nil
}
