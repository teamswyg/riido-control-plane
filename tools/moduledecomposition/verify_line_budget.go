package main

import "fmt"

func verifyLineBudgetConfig(budget fileLineBudget) error {
	if budget.TargetLines < 0 || budget.SampleLimit < 0 || budget.HotspotLimit < 0 ||
		budget.MaxFilesOverTarget < 0 || budget.MaxFileLines < 0 {
		return fmt.Errorf("file_line_budget values must be non-negative")
	}
	for _, limit := range budget.HotspotLimits {
		if limit.Path == "" || limit.MaxFiles < 0 || limit.MaxLines < 0 || limit.MaxTotalOver < 0 {
			return fmt.Errorf("file_line_budget hotspot limits must be complete and non-negative")
		}
	}
	return nil
}
