package main

import "fmt"

var mandatoryNextArtifacts = []string{
	"claim_binding",
	"verifier",
	"ci_gate",
	"redacted_evidence",
	"decision_record",
	"evidence_graph_edge",
}

func verifyRequiredNextArtifacts(source promotionSource) error {
	for _, artifact := range source.RequiredNextArtifacts {
		if !containsString(mandatoryNextArtifacts, artifact) {
			return fmt.Errorf("source %s declares unknown next artifact %s", source.ID, artifact)
		}
	}
	for _, artifact := range mandatoryNextArtifacts {
		if !containsString(source.RequiredNextArtifacts, artifact) {
			return fmt.Errorf("source %s missing required next artifact %s", source.ID, artifact)
		}
	}
	return nil
}

func containsString(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
