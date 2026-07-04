package main

import "fmt"

func verifyLoopCompletions(m manifest) error {
	items := loopCompletions(m)
	if len(items) == 0 {
		return fmt.Errorf("loop completion requires at least one loop")
	}
	for _, item := range items {
		if item.CompletionBasisPoints >= loopCompletionThresholdBasisPoints {
			continue
		}
		return fmt.Errorf("loop %s completion %d below threshold %d: %v",
			item.LoopID, item.CompletionBasisPoints,
			loopCompletionThresholdBasisPoints, item.MissingChecks)
	}
	return nil
}
