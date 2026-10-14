package main

import "fmt"

func refreshWorkflowCadenceMinutes(text string) (int, error) {
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

func expiryMinutes(loop loopRecord) int {
	return loop.ExpiresAfterHours * 60
}
