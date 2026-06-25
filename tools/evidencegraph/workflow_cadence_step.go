package main

import (
	"fmt"
	"strconv"
	"strings"
)

func positiveStepMinutes(expr, field string, unit int) (int, error) {
	step, err := strconv.Atoi(strings.TrimPrefix(field, "*/"))
	if err != nil || step <= 0 {
		return 0, fmt.Errorf("cron %q step must be positive", expr)
	}
	return step * unit, nil
}
