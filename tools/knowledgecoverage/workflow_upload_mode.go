package main

import "strings"

func workflowIfNoFilesFoundValue(line string) string {
	value, ok := strings.CutPrefix(strings.TrimSpace(line), "if-no-files-found:")
	if !ok {
		return ""
	}
	return strings.Trim(strings.TrimSpace(value), `"'`)
}
