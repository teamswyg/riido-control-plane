package main

import (
	"fmt"
	"strings"
)

const (
	architectureComponentDocSampleLimit      = 2
	architectureComponentDocSampleValueLimit = 96
)

func architectureComponentDocSample(values []string) string {
	if len(values) == 0 {
		return ""
	}
	limit := architectureComponentDocSampleLimit
	if len(values) < limit {
		limit = len(values)
	}
	parts := make([]string, 0, limit+1)
	for _, value := range values[:limit] {
		parts = append(parts, fmt.Sprintf("`%s`", markdownCellValue(value)))
	}
	if remaining := len(values) - limit; remaining > 0 {
		parts = append(parts, fmt.Sprintf("+%d", remaining))
	}
	return strings.Join(parts, "<br>")
}

func markdownCellValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\n", " ")
	if len(value) > architectureComponentDocSampleValueLimit {
		value = value[:architectureComponentDocSampleValueLimit] + "..."
	}
	value = strings.ReplaceAll(value, "|", "\\|")
	return value
}
