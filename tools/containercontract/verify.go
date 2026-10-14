package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/teamswyg/riido-control-plane/tools/containercontract/dockerfile"
)

type namedField struct{ Name, Value string }

func verifyContract(contract imageContract) (checkRecord, error) {
	parsed, err := dockerfile.Parse(contract.Dockerfile)
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

func verifyFinalStage(contract imageContract, parsed dockerfile.File) (*dockerfile.Stage, int, error) {
	finalStage := parsed.FinalStage()
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

func requireNamedFields(prefix string, fields []namedField) error {
	for _, field := range fields {
		if strings.TrimSpace(field.Value) == "" {
			return fmt.Errorf("%s%s is required", prefix, field.Name)
		}
	}
	return nil
}
