package main

import (
	"fmt"
	"os"
)

func captureRefreshCadence(root string, loop loopRecord, result *verifyResult) error {
	data, err := os.ReadFile(repoPath(root, loop.RefreshWorkflow))
	if err != nil {
		return fmt.Errorf("read refresh workflow %s: %w", loop.RefreshWorkflow, err)
	}
	minutes, err := refreshWorkflowCadenceMinutes(string(data))
	if err != nil {
		return fmt.Errorf("loop %s refresh workflow cadence: %w", loop.ID, err)
	}
	result.RefreshCadenceMinutes[loop.ID] = minutes
	return nil
}
