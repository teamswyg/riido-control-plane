package main

import (
	"errors"
	"fmt"
	"strings"
)

func validateContract(contract imageContract) error {
	if contract.SchemaVersion != contractSchemaVersion {
		return fmt.Errorf("schema_version must be %s", contractSchemaVersion)
	}
	if err := verifyEvidenceMetadata(contract.ID, contract.Assertions, contract.Loop); err != nil {
		return err
	}
	if err := requireContractFields(contract); err != nil {
		return err
	}
	return requireFinalContract(contract.Final)
}

func requireContractFields(contract imageContract) error {
	for field, value := range map[string]string{
		"service":                 contract.Service,
		"dockerfile":              contract.Dockerfile,
		"build.build_arg.name":    contract.Build.BuildArg.Name,
		"build.build_arg.default": contract.Build.BuildArg.Default,
		"build.stage_name":        contract.Build.StageName,
		"build.workdir":           contract.Build.Workdir,
		"build.cgo_enabled":       contract.Build.CGOEnabled,
		"build.go_build.package":  contract.Build.GoBuild.Package,
		"build.go_build.output":   contract.Build.GoBuild.Output,
		"final.base_image":        contract.Final.BaseImage,
		"final.copy_from":         contract.Final.CopyFrom,
		"final.copy_source":       contract.Final.CopySource,
		"final.binary":            contract.Final.Binary,
		"final.user":              contract.Final.User,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", field)
		}
	}
	return nil
}

func requireFinalContract(final finalContract) error {
	if len(final.ExposedPorts) == 0 {
		return errors.New("final.exposed_ports is required")
	}
	if len(final.Entrypoint) == 0 {
		return errors.New("final.entrypoint is required")
	}
	if len(final.Env) == 0 {
		return errors.New("final.env is required")
	}
	for i, copy := range final.RequiredCopies {
		if err := requireRequiredCopy(i, copy); err != nil {
			return err
		}
	}
	return nil
}

func requireRequiredCopy(i int, requiredCopy requiredCopyContract) error {
	if strings.TrimSpace(requiredCopy.From) == "" {
		return fmt.Errorf("final.required_copies[%d].from is required", i)
	}
	if strings.TrimSpace(requiredCopy.Source) == "" {
		return fmt.Errorf("final.required_copies[%d].source is required", i)
	}
	if strings.TrimSpace(requiredCopy.Destination) == "" {
		return fmt.Errorf("final.required_copies[%d].destination is required", i)
	}
	return nil
}
