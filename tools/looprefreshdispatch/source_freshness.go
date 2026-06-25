package main

import (
	"fmt"
	"strings"
	"time"
)

func verifySourceFresh(source refreshCommandEvidence, now time.Time) error {
	if strings.TrimSpace(source.GeneratedAt) == "" || strings.TrimSpace(source.ExpiresAt) == "" {
		return fmt.Errorf("refresh command evidence must carry generated_at and expires_at")
	}
	generatedAt, err := parseEvidenceTime(source.GeneratedAt, "source generated_at")
	if err != nil {
		return err
	}
	expiresAt, err := parseEvidenceTime(source.ExpiresAt, "source expires_at")
	if err != nil {
		return err
	}
	if !generatedAt.Before(expiresAt) {
		return fmt.Errorf("refresh command evidence generated_at must be before expires_at")
	}
	if now.Before(generatedAt) {
		return fmt.Errorf("refresh command evidence generated_at is in the future: %s", source.GeneratedAt)
	}
	if !now.Before(expiresAt) {
		return fmt.Errorf("refresh command evidence expired at %s", source.ExpiresAt)
	}
	return nil
}

func parseEvidenceTime(value, label string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, fmt.Errorf("%s: %w", label, err)
	}
	return parsed.UTC(), nil
}
