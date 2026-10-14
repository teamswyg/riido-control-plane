package main

import (
	"fmt"
	"os"
	"strings"
)

func verifyAll(repoRoot string, m manifest) (verifyResult, error) {
	if err := verifyManifestShape(m); err != nil {
		return verifyResult{}, err
	}
	if err := verifyLinkedCDManifest(repoRoot, m); err != nil {
		return verifyResult{}, err
	}
	result := verifyResult{BoundaryCount: len(m.Boundaries), RuleCount: len(m.Rules)}
	for _, item := range m.Boundaries {
		if err := verifyBoundary(repoRoot, item, &result); err != nil {
			return verifyResult{}, err
		}
	}
	return result, nil
}

func verifyManifestShape(m manifest) error {
	if m.SchemaVersion != manifestSchema {
		return fmt.Errorf("schema_version must be %s", manifestSchema)
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" || m.Workflow == "" {
		return fmt.Errorf("id, title, generated_doc, and workflow are required")
	}
	if m.Evidence == "" || m.LinkedCD == "" || len(m.Boundaries) == 0 || len(m.Rules) == 0 {
		return fmt.Errorf("evidence_artifact, linked manifest, boundaries, and rules are required")
	}
	return verifyLoop(m.Loop)
}

func verifyBoundary(repoRoot string, item boundary, result *verifyResult) error {
	if item.ID == "" || item.Owner == "" || item.Scope == "" || len(item.Evidence) == 0 {
		return fmt.Errorf("boundary id, owner, scope, and evidence are required")
	}
	for _, check := range item.Evidence {
		if err := verifyEvidenceCheck(repoRoot, item.ID, check, result); err != nil {
			return err
		}
	}
	return nil
}

func verifyEvidenceCheck(repoRoot, id string, check evidenceCheck, result *verifyResult) error {
	body, err := os.ReadFile(repoPath(repoRoot, check.Path))
	if err != nil {
		return fmt.Errorf("%s evidence %s: %w", id, check.Path, err)
	}
	result.EvidencePaths++
	text := string(body)
	for _, phrase := range check.Contains {
		result.PhraseChecks++
		if !strings.Contains(text, phrase) {
			return fmt.Errorf("%s evidence %s missing %q", id, check.Path, phrase)
		}
	}
	return nil
}
