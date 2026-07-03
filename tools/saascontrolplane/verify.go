package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/saascontrolplane/requirements"
	"github.com/teamswyg/riido-control-plane/tools/saascontrolplane/setutil"
)

func verifyManifest(repo string, m manifest) error {
	if m.SchemaVersion != requirements.ManifestSchema {
		return fmt.Errorf("schema_version = %q", m.SchemaVersion)
	}
	if m.ID != requirements.ExpectedID || m.Title == "" || m.GeneratedDoc == "" || m.Workflow == "" {
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
	set := setutil.StringSet(got)
	for _, contract := range requirements.RequiredSharedContracts {
		if !set[contract] {
			return fmt.Errorf("missing shared contract %q", contract)
		}
	}
	return nil
}
