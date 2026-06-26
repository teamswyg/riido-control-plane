package main

import (
	"fmt"
	"os"
	"time"
)

func verifySummaryFresh(summary liveSummary, now time.Time) error {
	generatedAt, err := parseEvidenceTime("generated_at", summary.GeneratedAt)
	if err != nil {
		return err
	}
	expiresAt, err := parseEvidenceTime("expires_at", summary.ExpiresAt)
	if err != nil {
		return err
	}
	if !expiresAt.After(generatedAt) {
		return fmt.Errorf("summary %s expires before it was generated", summary.ID)
	}
	if !now.UTC().Before(expiresAt) {
		return fmt.Errorf("summary %s expired at %s", summary.ID, summary.ExpiresAt)
	}
	return nil
}

func parseEvidenceTime(field, value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("summary %s is required", field)
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("summary %s is invalid: %w", field, err)
	}
	return parsed.UTC(), nil
}

func promotionNow() time.Time {
	if value := os.Getenv("RIIDO_HARNESS_PROMOTION_NOW"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err == nil {
			return parsed.UTC()
		}
	}
	return time.Now().UTC()
}
