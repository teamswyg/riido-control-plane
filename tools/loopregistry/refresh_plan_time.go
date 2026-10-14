package main

import "time"

func refreshEvidenceGeneratedAt(generatedAt time.Time) string {
	if generatedAt.IsZero() {
		return ""
	}
	return formatEvidenceTime(generatedAt)
}

func refreshDueAt(generatedAt time.Time, cadenceMinutes int) string {
	if generatedAt.IsZero() || cadenceMinutes == 0 {
		return ""
	}
	return formatEvidenceTime(generatedAt.Add(time.Duration(cadenceMinutes) * time.Minute))
}

func refreshExpiresAt(generatedAt time.Time, expiresAfterHours int) string {
	if generatedAt.IsZero() || expiresAfterHours == 0 {
		return ""
	}
	return formatEvidenceTime(generatedAt.Add(time.Duration(expiresAfterHours) * time.Hour))
}
