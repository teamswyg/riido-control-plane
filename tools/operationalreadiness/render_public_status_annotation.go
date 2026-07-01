package main

import (
	"fmt"
	"strings"
)

func renderPublicStatusGitHubAnnotation(status publicStatus) string {
	level := "notice"
	if status.Overall != "operational" {
		level = "warning"
	}
	message := fmt.Sprintf(
		"overall=%s visibility=%s generated_at=%s expires_at=%s run_id=%s p0_partial=%d partial=%d stale=%d categories=%s candidates=%d next=%s",
		status.Overall,
		status.Visibility,
		status.GeneratedAt,
		status.ExpiresAt,
		status.SourceRunID,
		status.P0PartialCount,
		status.PartialCount,
		status.StalePartialCount,
		publicBlockingCategorySummary(status.BlockingCategories),
		status.ClosedLoopCandidates,
		status.NextArtifact,
	)
	return fmt.Sprintf(
		"::%s title=%s::%s\n",
		level,
		githubAnnotationEscape("Riido Public QA Status"),
		githubAnnotationEscape(message),
	)
}

func githubAnnotationEscape(value string) string {
	value = strings.ReplaceAll(value, "%", "%25")
	value = strings.ReplaceAll(value, "\r", "%0D")
	value = strings.ReplaceAll(value, "\n", "%0A")
	value = strings.ReplaceAll(value, ":", "%3A")
	value = strings.ReplaceAll(value, ",", "%2C")
	return value
}

func publicBlockingCategorySummary(categories []publicStatusCategory) string {
	if len(categories) == 0 {
		return "none"
	}
	values := make([]string, 0, len(categories))
	for _, category := range categories {
		values = append(values, category.Category)
	}
	return strings.Join(values, "|")
}
