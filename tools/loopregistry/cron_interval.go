package main

import (
	"fmt"
	"strings"
)

const minutesPerDay = 24 * 60

func cronIntervalMinutes(expr string) (int, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return 0, fmt.Errorf("cron %q must have five fields", expr)
	}
	minute, hour := fields[0], fields[1]
	if fields[2] != "*" || fields[3] != "*" || fields[4] != "*" {
		return 0, fmt.Errorf("cron %q must use daily-or-more-frequent wildcard date fields", expr)
	}
	if hour == "*" {
		return hourlyCronMinutes(expr, minute)
	}
	if strings.HasPrefix(hour, "*/") {
		return steppedHoursCronMinutes(expr, minute, hour)
	}
	if isInt(minute) && isInt(hour) {
		return minutesPerDay, nil
	}
	return 0, fmt.Errorf("cron %q uses unsupported minute/hour cadence", expr)
}

func hourlyCronMinutes(expr, minute string) (int, error) {
	if minute == "*" {
		return 1, nil
	}
	if strings.HasPrefix(minute, "*/") {
		return stepMinutes(expr, minute)
	}
	if isInt(minute) {
		return 60, nil
	}
	return 0, fmt.Errorf("cron %q uses unsupported minute cadence", expr)
}
