package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type verificationResult struct {
	LocalEvidence     int
	ExternalEvidence  int
	RemainingBoundary int
}

func verifyManifest(repoRoot string, manifest evidenceManifest) (verificationResult, error) {
	docPath := filepath.Join(repoRoot, filepath.FromSlash(manifest.HumanDoc))
	docBytes, err := os.ReadFile(docPath)
	if err != nil {
		return verificationResult{}, fmt.Errorf("read human doc: %w", err)
	}
	doc := string(docBytes)
	if err := verifyHeader(manifest, doc); err != nil {
		return verificationResult{}, err
	}
	seen, err := verifyEvidence(repoRoot, manifest, doc)
	if err != nil {
		return verificationResult{}, err
	}
	if err := verifyRequiredRisks(seen); err != nil {
		return verificationResult{}, err
	}
	if err := verifyBoundaries(manifest.RemainingBoundary); err != nil {
		return verificationResult{}, err
	}
	return verificationResult{len(manifest.LocalEvidence), len(manifest.ExternalEvidence), len(manifest.RemainingBoundary)}, nil
}

func verifyRequiredRisks(seen map[string]bool) error {
	for _, risk := range requiredRisks {
		if !seen[risk] {
			return fmt.Errorf("manifest missing risk evidence %q", risk)
		}
	}
	return nil
}

func docMentions(doc, value string) bool {
	return strings.Contains(doc, value)
}

func expectedRisk(risk string) bool {
	return slices.Contains(requiredRisks, risk)
}
