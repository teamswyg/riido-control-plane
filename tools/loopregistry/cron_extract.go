package main

import (
	"strings"
)

func workflowCronExpressions(text string) []string {
	expressions := []string{}
	for _, line := range strings.Split(text, "\n") {
		cron, ok := cronExpressionFromLine(line)
		if ok {
			expressions = append(expressions, cron)
		}
	}
	return expressions
}

func cronExpressionFromLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "- cron:") {
		return "", false
	}
	value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- cron:"))
	value = strings.Trim(value, `"'`)
	return value, value != ""
}
