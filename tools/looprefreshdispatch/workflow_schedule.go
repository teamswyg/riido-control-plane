package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func workflowDailyCronMinute(root, workflow string) (int, error) {
	data, err := os.ReadFile(repoPath(root, ".github/workflows/"+workflow))
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		expr, ok := cronExpressionFromLine(line)
		if !ok {
			continue
		}
		return dailyCronMinute(expr)
	}
	return 0, fmt.Errorf("workflow %s has no cron schedule", workflow)
}

func requireRefreshDispatchAfterProducer(root string) error {
	producer, err := workflowDailyCronMinute(root, "loop-registry.yml")
	if err != nil {
		return err
	}
	dispatcher, err := workflowDailyCronMinute(root, "loop-refresh-dispatch.yml")
	if err != nil {
		return err
	}
	return requireRefreshDispatchOrder(producer, dispatcher)
}

func requireRefreshDispatchOrder(producer, dispatcher int) error {
	if dispatcher <= producer {
		return fmt.Errorf(
			"loop-refresh-dispatch must run after loop-registry: producer=%d dispatcher=%d",
			producer,
			dispatcher,
		)
	}
	return nil
}

func cronExpressionFromLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "- cron:") {
		return "", false
	}
	return strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "- cron:")), `"'`), true
}

func dailyCronMinute(expr string) (int, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 || fields[2] != "*" || fields[3] != "*" || fields[4] != "*" {
		return 0, fmt.Errorf("unsupported cron %q", expr)
	}
	minute, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, fmt.Errorf("cron minute: %w", err)
	}
	hour, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, fmt.Errorf("cron hour: %w", err)
	}
	return hour*60 + minute, nil
}
