package main

import (
	"fmt"
	"time"
)

func verifyCandidateFresh(candidate candidateEvidence, now time.Time) error {
	generatedAt, err := parseCandidateTime("source_generated_at", candidate.SourceGeneratedAt)
	if err != nil {
		return err
	}
	expiresAt, err := parseCandidateTime("source_expires_at", candidate.SourceExpiresAt)
	if err != nil {
		return err
	}
	if !expiresAt.After(generatedAt) {
		return fmt.Errorf("candidate artifact %s source_expires_at must be after source_generated_at", candidate.ID)
	}
	if !now.Before(expiresAt) {
		return fmt.Errorf("candidate artifact %s expired at %s", candidate.ID, candidate.SourceExpiresAt)
	}
	return nil
}

func parseCandidateTime(field, value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("candidate artifact %s is required", field)
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("candidate artifact %s is invalid: %w", field, err)
	}
	return parsed.UTC(), nil
}
