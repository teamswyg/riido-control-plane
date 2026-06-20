package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const evidenceSchemaVersion = "riido-dependency-allowlist-evidence.v1"

type evidence struct {
	SchemaVersion              string `json:"schema_version"`
	Service                    string `json:"service"`
	Status                     string `json:"status"`
	DirectDependenciesVerified int    `json:"direct_dependencies_verified"`
	AllowedDirectModules       int    `json:"allowed_direct_modules"`
}

func newEvidence(c contract, report dependencyReport) evidence {
	return evidence{
		SchemaVersion:              evidenceSchemaVersion,
		Service:                    c.Service,
		Status:                     "verified",
		DirectDependenciesVerified: report.DirectDependenciesVerified,
		AllowedDirectModules:       report.AllowedDirectModules,
	}
}

func writeEvidence(path string, value evidence) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create evidence dir: %w", err)
	}
	if err := os.WriteFile(path, append(body, '\n'), 0o644); err != nil {
		return fmt.Errorf("write evidence: %w", err)
	}
	return nil
}
