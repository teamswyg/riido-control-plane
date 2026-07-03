package main

import (
	"fmt"
	"os"

	"github.com/teamswyg/riido-control-plane/tools/apiclientdelivery/pathutil"
	"github.com/teamswyg/riido-control-plane/tools/apiclientdelivery/requirements"
)

func verifyAll(repoRoot string, m manifest) (verifyResult, error) {
	if err := verifyShape(m); err != nil {
		return verifyResult{}, err
	}
	result := verifyResult{
		SourceManifests: len(m.Sources), Owners: len(m.Owners),
		FigmaContexts: len(m.Figma), SourceChecks: len(m.Checks),
	}
	if err := verifySourceManifests(repoRoot, m.Sources); err != nil {
		return verifyResult{}, err
	}
	if err := verifySourceChecks(repoRoot, m.Checks, &result); err != nil {
		return verifyResult{}, err
	}
	if err := collectRiskEvidenceTests(repoRoot, m.RiskEvidence, &result); err != nil {
		return verifyResult{}, err
	}
	if err := verifyRenderedPhrases(m, renderDoc(m, result), &result); err != nil {
		return verifyResult{}, err
	}
	return result, nil
}

func verifyShape(m manifest) error {
	if m.SchemaVersion != requirements.ManifestSchema {
		return fmt.Errorf("schema_version must be %s", requirements.ManifestSchema)
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" || m.Workflow == "" || m.RiskEvidence == "" {
		return fmt.Errorf("id, title, generated_doc, workflow, and risk_evidence_manifest are required")
	}
	if len(m.Sources) == 0 || len(m.Owners) == 0 || len(m.Figma) == 0 || len(m.Checks) == 0 {
		return fmt.Errorf("sources, owners, figma contexts, and source checks are required")
	}
	return verifyLoop(m.Loop)
}

func verifySourceManifests(repoRoot string, sources []sourceRef) error {
	for _, source := range sources {
		if source.Name == "" || source.Path == "" {
			return fmt.Errorf("source manifest name and path are required")
		}
		if _, err := os.Stat(pathutil.Resolve(repoRoot, source.Path)); err != nil {
			return fmt.Errorf("missing source manifest %s: %w", source.Path, err)
		}
	}
	return nil
}
