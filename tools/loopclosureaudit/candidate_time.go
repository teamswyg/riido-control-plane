package main

import (
	"os"
	"time"
)

const candidateTimeLayout = time.RFC3339

func candidateNow() time.Time {
	value := os.Getenv("RIIDO_LOOP_CLOSURE_AUDIT_NOW")
	if value == "" {
		return time.Now().UTC()
	}
	parsed, err := time.Parse(candidateTimeLayout, value)
	if err != nil {
		return time.Now().UTC()
	}
	return parsed.UTC()
}

func candidateWindow() (string, string) {
	generatedAt := candidateNow()
	expiresAt := generatedAt.Add(24 * time.Hour)
	return formatCandidateTime(generatedAt), formatCandidateTime(expiresAt)
}

func formatCandidateTime(value time.Time) string {
	return value.UTC().Format(candidateTimeLayout)
}
