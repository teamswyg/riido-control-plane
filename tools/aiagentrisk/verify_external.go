package main

import (
	"fmt"
	"strings"
)

func verifyExternalEvidence(evidence externalEvidence, doc string) error {
	if evidence.Risk == "" || evidence.Status != "verified" || evidence.Proves == "" {
		return fmt.Errorf("invalid external evidence %+v", evidence)
	}
	if !expectedRisk(evidence.Risk) {
		return fmt.Errorf("unexpected external evidence risk %q", evidence.Risk)
	}
	if evidence.Repo != "riido-contracts" {
		return fmt.Errorf("external evidence repo must stay at repo boundary, got %q", evidence.Repo)
	}
	if strings.Contains(evidence.Test, "/") || strings.Contains(evidence.Test, "internal/") {
		return fmt.Errorf("external evidence must not reference private package paths: %+v", evidence)
	}
	if !docMentions(doc, evidence.Test) {
		return fmt.Errorf("human doc must mention external evidence test %s", evidence.Test)
	}
	return nil
}
