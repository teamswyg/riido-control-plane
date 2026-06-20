package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type riskManifest struct {
	Local    []riskEvidence `json:"local_evidence"`
	External []riskEvidence `json:"external_evidence"`
}

type riskEvidence struct {
	Risk   string `json:"risk"`
	Test   string `json:"test"`
	Proves string `json:"proves"`
}

func collectRiskEvidenceTests(repoRoot, path string, result *verifyResult) error {
	body, err := os.ReadFile(repoPath(repoRoot, path))
	if err != nil {
		return fmt.Errorf("read risk evidence manifest: %w", err)
	}
	var manifest riskManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return fmt.Errorf("decode risk evidence manifest: %w", err)
	}
	result.RiskEvidence = append(append([]riskEvidence{}, manifest.Local...), manifest.External...)
	result.RiskTests = len(result.RiskEvidence)
	return nil
}
