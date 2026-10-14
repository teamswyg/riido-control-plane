package main

import (
	"os"
	"time"
)

const candidateDecisionEvidenceTTLHours = 24

func evidenceWindow(ttlHours int) (string, string) {
	now := evidenceNow()
	return formatEvidenceTime(now), formatEvidenceTime(now.Add(time.Duration(ttlHours) * time.Hour))
}

func evidenceNow() time.Time {
	if value := os.Getenv("RIIDO_EVIDENCE_NOW"); value != "" {
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			return parsed.UTC()
		}
	}
	return time.Now().UTC()
}

func formatEvidenceTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
