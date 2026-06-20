package main

import (
	"fmt"
	"slices"
	"strings"
)

type verifyResult struct {
	RuntimeEnvCount int
	ManifestCount   int
	SecretCount     int
	OperatorCount   int
}

func verifyAll(repoRoot string, m manifest) (verifyResult, error) {
	if err := verifyManifestShape(m); err != nil {
		return verifyResult{}, err
	}
	sourceNames, err := scanRuntimeEnv(repoRoot, m.SourceDir)
	if err != nil {
		return verifyResult{}, err
	}
	if err := verifyEnvParity(m, sourceNames); err != nil {
		return verifyResult{}, err
	}
	result := verifyResult{RuntimeEnvCount: len(sourceNames), ManifestCount: len(m.Entries)}
	for _, entry := range m.Entries {
		if strings.Contains(entry.Sensitivity, "secret") || entry.Sensitivity == "credential" {
			result.SecretCount++
		}
		if entry.Sensitivity == "operator" {
			result.OperatorCount++
		}
	}
	return result, nil
}

func verifyManifestShape(m manifest) error {
	if m.SchemaVersion == "" || m.ID == "" || m.Title == "" || m.GeneratedDoc == "" {
		return fmt.Errorf("schema_version, id, title, and generated_doc are required")
	}
	if m.Workflow == "" || m.Evidence == "" || m.SourceDir == "" || len(m.Entries) == 0 {
		return fmt.Errorf("workflow, evidence_artifact, source_dir, and entries are required")
	}
	return verifyLoop(m.Loop)
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
