package main

import (
	"errors"
	"fmt"
	"strings"
)

func verifyContract(contract imageContract) (checkRecord, error) {
	parsed, err := parseDockerfile(contract.Dockerfile)
	if err != nil {
		return checkRecord{}, err
	}
	checks, buildStage, err := verifyBuildStage(contract, parsed)
	if err != nil {
		return checkRecord{}, err
	}
	finalStage, finalChecks, err := verifyFinalStage(contract, parsed)
	if err != nil {
		return checkRecord{}, err
	}
	checks += finalChecks
	if strings.TrimSpace(contract.FargateTaskDefinitionIR) != "" {
		if err := verifyTaskDefinitionIR(contract); err != nil {
			return checkRecord{}, err
		}
		checks += 3
	}
	return newCheckRecord(contract, buildStage, finalStage, checks), nil
}

func verifyBuildStage(contract imageContract, parsed dockerfile) (int, *stage, error) {
	checks := 0
	require := countedRequire(&checks)
	if err := require(parsed.Args[contract.Build.BuildArg.Name] == contract.Build.BuildArg.Default,
		"ARG %s default = %q, want %q", contract.Build.BuildArg.Name, parsed.Args[contract.Build.BuildArg.Name], contract.Build.BuildArg.Default); err != nil {
		return checks, nil, err
	}
	buildStage := parsed.stageByAlias(contract.Build.StageName)
	if buildStage == nil {
		return checks, nil, fmt.Errorf("build stage %q not found", contract.Build.StageName)
	}
	if err := require(buildStage.Base == "${"+contract.Build.BuildArg.Name+"}", "build stage base = %q, want ${%s}", buildStage.Base, contract.Build.BuildArg.Name); err != nil {
		return checks, nil, err
	}
	if err := require(buildStage.Workdir == contract.Build.Workdir, "build WORKDIR = %q, want %q", buildStage.Workdir, contract.Build.Workdir); err != nil {
		return checks, nil, err
	}
	if err := require(buildStage.Env["CGO_ENABLED"] == contract.Build.CGOEnabled, "CGO_ENABLED = %q, want %q", buildStage.Env["CGO_ENABLED"], contract.Build.CGOEnabled); err != nil {
		return checks, nil, err
	}
	if err := require(hasGoBuildRun(buildStage.Runs, contract.Build.GoBuild), "go build command does not satisfy container contract"); err != nil {
		return checks, nil, err
	}
	return checks, buildStage, nil
}

func verifyFinalStage(contract imageContract, parsed dockerfile) (*stage, int, error) {
	finalStage := parsed.finalStage()
	if finalStage == nil {
		return nil, 0, errors.New("final stage not found")
	}
	checks := 0
	require := countedRequire(&checks)
	if err := require(finalStage.Base == contract.Final.BaseImage, "final base image = %q, want %q", finalStage.Base, contract.Final.BaseImage); err != nil {
		return nil, checks, err
	}
	if err := verifyFinalStageMore(contract, finalStage, require); err != nil {
		return nil, checks, err
	}
	return finalStage, checks, nil
}
