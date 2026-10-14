package main

import (
	"os"
	"time"
)

func liveEvidenceNow() time.Time {
	if value := os.Getenv("RIIDO_LIVE_WORKFLOW_EVIDENCE_NOW"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err == nil {
			return parsed.UTC()
		}
	}
	return time.Now().UTC()
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
