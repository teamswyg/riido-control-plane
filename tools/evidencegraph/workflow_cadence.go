package main

import (
	"fmt"
	"strings"
)

func refreshCadenceMinutes(text string) (int, error) {
	expressions := workflowCronExpressions(text)
	if len(expressions) == 0 {
		return 0, fmt.Errorf("refresh workflow has no cron schedule")
	}
	best := 0
	for _, expr := range expressions {
		minutes, err := cronIntervalMinutes(expr)
		if err != nil {
			return 0, err
		}
		if best == 0 || minutes < best {
			best = minutes
		}
	}
	return best, nil
}

func workflowCronExpressions(text string) []string {
	out := []string{}
	for _, line := range strings.Split(text, "\n") {
		if cron, ok := cronExpressionFromLine(line); ok {
			out = append(out, cron)
		}
	}
	return out
}

func cronExpressionFromLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "- cron:") {
		return "", false
	}
	value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- cron:"))
	return strings.Trim(value, `"'`), true
}

func cronIntervalMinutes(expr string) (int, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return 0, fmt.Errorf("cron %q must have five fields", expr)
	}
	if fields[2] != "*" || fields[3] != "*" || fields[4] != "*" {
		return 0, fmt.Errorf("cron %q must use daily-or-more-frequent wildcard date fields", expr)
	}
	return minuteHourInterval(expr, fields[0], fields[1])
}

func minuteHourInterval(expr, minute, hour string) (int, error) {
	if strings.HasPrefix(minute, "*/") && hour == "*" {
		return positiveStepMinutes(expr, minute, 1)
	}
	if minute != "*" && strings.HasPrefix(hour, "*/") {
		return positiveStepMinutes(expr, hour, 60)
	}
	if minute != "*" && hour != "*" {
		return 24 * 60, nil
	}
	return 0, fmt.Errorf("cron %q uses unsupported minute/hour cadence", expr)
}
