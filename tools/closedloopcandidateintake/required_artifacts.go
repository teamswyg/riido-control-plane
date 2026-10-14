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

func verifyRequiredNextArtifacts(values []string, sourceID string) error {
	for _, value := range values {
		if !containsString(mandatoryNextArtifacts, value) {
			return fmt.Errorf("source %s declares unknown next artifact %s", sourceID, value)
		}
	}
	for _, value := range mandatoryNextArtifacts {
		if !containsString(values, value) {
			return fmt.Errorf("source %s missing required next artifact %s", sourceID, value)
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
