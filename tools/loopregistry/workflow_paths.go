package main

import "strings"

func workflowPathFilters(text string) map[string]bool {
	paths := map[string]bool{}
	lines := strings.Split(text, "\n")
	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "paths:" {
			continue
		}
		for i++; i < len(lines); i++ {
			line := lines[i]
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if !strings.HasPrefix(trimmed, "- ") {
				break
			}
			paths[strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")), `"'`)] = true
		}
	}
	return paths
}
