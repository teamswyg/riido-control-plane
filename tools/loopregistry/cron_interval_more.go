package main

import (
	"fmt"
	"strconv"
	"strings"
)

func steppedHoursCronMinutes(expr, minute, hour string) (int, error) {
	if !isInt(minute) {
		return 0, fmt.Errorf("cron %q uses stepped hours without fixed minute", expr)
	}
	step, err := stepValue(hour)
	if err != nil {
		return 0, fmt.Errorf("cron %q hour step: %w", expr, err)
	}
	return step * 60, nil
}

func stepMinutes(expr, minute string) (int, error) {
	step, err := stepValue(minute)
	if err != nil {
		return 0, fmt.Errorf("cron %q minute step: %w", expr, err)
	}
	return step, nil
}

func stepValue(value string) (int, error) {
	step, err := strconv.Atoi(strings.TrimPrefix(value, "*/"))
	if err != nil || step <= 0 {
		return 0, fmt.Errorf("invalid step %q", value)
	}
	return step, nil
}

func isInt(value string) bool {
	_, err := strconv.Atoi(value)
	return err == nil
}
