package main

import (
	"fmt"

	"github.com/teamswyg/riido-control-plane/tools/aiagentclientapi/doccheck"
	"github.com/teamswyg/riido-control-plane/tools/aiagentclientapi/pathutil"
	"github.com/teamswyg/riido-control-plane/tools/aiagentclientapi/requirements"
	"github.com/teamswyg/riido-control-plane/tools/aiagentclientapi/setutil"
)

func verify(root string, m manifest, checkDoc bool) error {
	if err := verifyHeader(m); err != nil {
		return err
	}
	if err := verifyContractMirror(root, m); err != nil {
		return err
	}
	if err := verifyStaticLists(m); err != nil {
		return err
	}
	if err := verifyThreadHistoryV3(m.ThreadHistoryV3); err != nil {
		return err
	}
	if err := verifyDeviceClientGuidance(m.DeviceClientGuidance); err != nil {
		return err
	}
	if err := verifySources(root, m.SourceChecks); err != nil {
		return err
	}
	if err := verifyLoop(m.Loop); err != nil {
		return err
	}
	if checkDoc {
		return doccheck.Verify(pathutil.Resolve(root, m.GeneratedDoc), renderDoc(m))
	}
	return nil
}

func verifyHeader(m manifest) error {
	if m.SchemaVersion != requirements.ManifestSchema || m.ID != requirements.ExpectedID || m.RiidoTask != requirements.ExpectedTask {
		return fmt.Errorf("unexpected manifest identity")
	}
	required := []string{m.Title, m.GeneratedDoc, m.Workflow, m.EvidenceArtifact}
	for _, value := range required {
		if value == "" {
			return fmt.Errorf("title, generated_doc, workflow, and evidence_artifact are required")
		}
	}
	return nil
}

func verifyStaticLists(m manifest) error {
	if err := setutil.RequireStrings("runtime config", m.RuntimeConfigKeys, requirements.RequiredRuntimeConfigKeys); err != nil {
		return err
	}
	if err := setutil.RequireStrings("public field", m.PublicFields, requirements.RequiredPublicFields); err != nil {
		return err
	}
	return setutil.RequireStrings("deployment evidence", m.DeploymentEvidence, requirements.RequiredDeploymentEvidence)
}
