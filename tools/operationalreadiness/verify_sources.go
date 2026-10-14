package main

import "fmt"

func verifyCandidateSources(m manifest) error {
	source := readinessCandidateSource(m)
	if source.ID != readinessCandidateSourceID ||
		source.SourceWorkflow != m.Workflow ||
		source.CandidateArtifact != readinessCandidateArtifact ||
		source.HarnessLoop != readinessHarnessLoop ||
		source.PromotionTarget != readinessPromotionTarget {
		return fmt.Errorf("operational readiness candidate source drift")
	}
	return verifyRequiredCandidateArtifacts(source.RequiredNextArtifacts)
}

func verifyRequiredCandidateArtifacts(values []string) error {
	required := candidateRequiredArtifacts()
	for _, value := range required {
		if !containsString(values, value) {
			return fmt.Errorf("candidate source missing required artifact %s", value)
		}
	}
	for _, value := range values {
		if !containsString(required, value) {
			return fmt.Errorf("candidate source has unknown artifact %s", value)
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
