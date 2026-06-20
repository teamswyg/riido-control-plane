package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const evidenceSchemaVersion = "riido-ai-agent-risk-evidence-result.v1"

type evidenceResult struct {
	SchemaVersion     string `json:"schema_version"`
	ID                string `json:"id"`
	Status            string `json:"status"`
	LocalEvidence     int    `json:"local_evidence"`
	ExternalEvidence  int    `json:"external_evidence"`
	RemainingBoundary int    `json:"remaining_boundaries"`
}

func newEvidence(manifest evidenceManifest, result verificationResult) evidenceResult {
	return evidenceResult{
		SchemaVersion:     evidenceSchemaVersion,
		ID:                manifest.ID,
		Status:            "verified",
		LocalEvidence:     result.LocalEvidence,
		ExternalEvidence:  result.ExternalEvidence,
		RemainingBoundary: result.RemainingBoundary,
	}
}

func writeEvidence(path string, value evidenceResult) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create evidence dir: %w", err)
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}
