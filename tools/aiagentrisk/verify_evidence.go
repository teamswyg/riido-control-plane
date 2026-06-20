package main

import (
	"fmt"
	"strings"
)

func verifyEvidence(repoRoot string, manifest evidenceManifest, doc string) (map[string]bool, error) {
	seen := map[string]bool{}
	for _, evidence := range manifest.LocalEvidence {
		if err := verifyLocalEvidence(repoRoot, evidence, doc); err != nil {
			return nil, err
		}
		seen[evidence.Risk] = true
	}
	for _, evidence := range manifest.ExternalEvidence {
		if err := verifyExternalEvidence(evidence, doc); err != nil {
			return nil, err
		}
		seen[evidence.Risk] = true
	}
	return seen, nil
}

func verifyLocalEvidence(repoRoot string, evidence localEvidence, doc string) error {
	if evidence.Risk == "" || evidence.Status != "verified" || evidence.Proves == "" {
		return fmt.Errorf("invalid local evidence %+v", evidence)
	}
	if !expectedRisk(evidence.Risk) || !strings.HasPrefix(evidence.Test, "Test") {
		return fmt.Errorf("invalid local evidence risk/test %+v", evidence)
	}
	found, err := testFunctionExists(repoRoot, evidence.Package, evidence.Test)
	if err != nil || !found {
		return fmt.Errorf("%s does not contain %s: %w", evidence.Package, evidence.Test, err)
	}
	if !docMentions(doc, evidence.Test) {
		return fmt.Errorf("human doc must mention evidence test %s", evidence.Test)
	}
	return nil
}
