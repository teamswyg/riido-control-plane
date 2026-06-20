package main

import (
	"fmt"
)

func verifyManifest(repo string, m manifest) error {
	if m.SchemaVersion != manifestSchema {
		return fmt.Errorf("schema_version = %q", m.SchemaVersion)
	}
	if m.ID != expectedID || m.Title == "" || m.GeneratedDoc == "" || m.Workflow == "" {
		return fmt.Errorf("manifest identity is incomplete")
	}
	if m.EvidenceArtifact == "" || m.OwnerPackage == "" {
		return fmt.Errorf("evidence_artifact and owner_package are required")
	}
	if err := verifySharedContracts(m.SharedContracts); err != nil {
		return err
	}
	if err := verifyFocusedWorkflows(repo, m.FocusedWorkflows); err != nil {
		return err
	}
	if err := verifyBoundaries(repo, m); err != nil {
		return err
	}
	return verifyLoop(m.Loop)
}

func verifySharedContracts(got []string) error {
	set := stringSet(got)
	for _, contract := range requiredSharedContracts {
		if !set[contract] {
			return fmt.Errorf("missing shared contract %q", contract)
		}
	}
	return nil
}
