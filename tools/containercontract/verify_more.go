package main

import (
	"fmt"
	"reflect"
	"strings"
)

type requireFunc func(bool, string, ...any) error

func countedRequire(checks *int) requireFunc {
	return func(ok bool, format string, args ...any) error {
		*checks++
		if !ok {
			return fmt.Errorf(format, args...)
		}
		return nil
	}
}

func verifyFinalStageMore(contract imageContract, finalStage *stage, require requireFunc) error {
	if err := require(finalStage.User == contract.Final.User, "final USER = %q, want %q", finalStage.User, contract.Final.User); err != nil {
		return err
	}
	if err := require(reflect.DeepEqual(finalStage.Entrypoint, contract.Final.Entrypoint), "ENTRYPOINT = %v, want %v", finalStage.Entrypoint, contract.Final.Entrypoint); err != nil {
		return err
	}
	if err := require(intSetEqual(finalStage.Exposes, contract.Final.ExposedPorts), "EXPOSE = %v, want %v", finalStage.Exposes, contract.Final.ExposedPorts); err != nil {
		return err
	}
	return verifyFinalStageCopies(contract, finalStage, require)
}

func verifyFinalStageCopies(contract imageContract, finalStage *stage, require requireFunc) error {
	for key, want := range contract.Final.Env {
		if err := require(finalStage.Env[key] == want, "final ENV %s = %q, want %q", key, finalStage.Env[key], want); err != nil {
			return err
		}
	}
	if err := require(hasCopy(finalStage.Copies, contract.Final.CopyFrom, contract.Final.CopySource, contract.Final.Binary), "final COPY from %s %s -> %s not found", contract.Final.CopyFrom, contract.Final.CopySource, contract.Final.Binary); err != nil {
		return err
	}
	for _, requiredCopy := range contract.Final.RequiredCopies {
		if err := require(hasCopy(finalStage.Copies, requiredCopy.From, requiredCopy.Source, requiredCopy.Destination), "final required COPY from %s %s -> %s not found", requiredCopy.From, requiredCopy.Source, requiredCopy.Destination); err != nil {
			return err
		}
	}
	if err := require(contract.Final.User != "0" && !strings.HasPrefix(contract.Final.User, "0:"), "final user must be non-root"); err != nil {
		return err
	}
	return nil
}
