package main

import (
	"errors"
	"fmt"
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

func requireContractFields(c imageContract) error {
	return requireNamedFields("", []namedField{
		{"service", c.Service},
		{"dockerfile", c.Dockerfile},
		{"build.build_arg.name", c.Build.BuildArg.Name},
		{"build.build_arg.default", c.Build.BuildArg.Default},
		{"build.stage_name", c.Build.StageName},
		{"build.workdir", c.Build.Workdir},
		{"build.cgo_enabled", c.Build.CGOEnabled},
		{"build.module_download", c.Build.ModuleDownload.Command},
		{"build.go_build.package", c.Build.GoBuild.Package},
		{"build.go_build.output", c.Build.GoBuild.Output},
		{"final.base_image", c.Final.BaseImage},
		{"final.copy_from", c.Final.CopyFrom},
		{"final.copy_source", c.Final.CopySource},
		{"final.binary", c.Final.Binary},
		{"final.user", c.Final.User},
	})
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
	for i, requiredCopy := range final.RequiredCopies {
		if err := requireRequiredCopy(i, requiredCopy); err != nil {
			return err
		}
	}
	return nil
}

func requireRequiredCopy(i int, requiredCopy requiredCopyContract) error {
	prefix := fmt.Sprintf("final.required_copies[%d].", i)
	return requireNamedFields(prefix, []namedField{
		{"from", requiredCopy.From},
		{"source", requiredCopy.Source},
		{"destination", requiredCopy.Destination},
	})
}
