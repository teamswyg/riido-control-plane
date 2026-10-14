package main

import "fmt"

func verifyCandidateRequiredNextArtifacts(item closedLoopCandidate, sourceID string) error {
	allowed, err := candidateAllowedNextArtifacts(item)
	if err != nil {
		return err
	}
	for _, value := range item.RequiredNextArtifacts {
		if !containsString(allowed, value) {
			return fmt.Errorf("source %s declares unknown next artifact %s", sourceID, value)
		}
	}
	for _, value := range mandatoryNextArtifacts {
		if !containsString(item.RequiredNextArtifacts, value) {
			return fmt.Errorf("source %s missing required next artifact %s", sourceID, value)
		}
	}
	return nil
}

func candidateAllowedNextArtifacts(item closedLoopCandidate) ([]string, error) {
	allowed := append([]string(nil), mandatoryNextArtifacts...)
	next, err := subjectNextArtifact(item)
	if err != nil || next == "" || containsString(allowed, next) {
		return allowed, err
	}
	return append(allowed, next), nil
}
