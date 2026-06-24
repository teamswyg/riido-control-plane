package main

import "fmt"

func verifyLoops(root string, m manifest) (map[string]bool, verifyResult, error) {
	ids := map[string]bool{}
	result := verifyResult{RefreshCadenceMinutes: map[string]int{}}
	for _, loop := range m.Loops {
		if ids[loop.ID] || loop.ID == "" {
			return nil, result, fmt.Errorf("loop id must be unique and non-empty: %q", loop.ID)
		}
		ids[loop.ID] = true
		if err := verifyLoop(root, m, loop); err != nil {
			return nil, result, err
		}
		if err := captureRefreshCadence(root, loop, &result); err != nil {
			return nil, result, err
		}
		result.Loops++
		if loop.Kind == kindHarness {
			result.Harnesses++
		}
		if loop.Kind == kindClosedLoop {
			result.ClosedLoops++
		}
		if loop.ExpiresAfterHours > result.MaxExpiryHours {
			result.MaxExpiryHours = loop.ExpiresAfterHours
		}
	}
	return ids, result, nil
}

func verifyLoop(root string, m manifest, loop loopRecord) error {
	if loop.Kind != kindHarness && loop.Kind != kindClosedLoop {
		return fmt.Errorf("loop %s has unsupported kind %q", loop.ID, loop.Kind)
	}
	if len(loop.Observes) == 0 || len(loop.Verifies) == 0 || len(loop.Evidence) == 0 {
		return fmt.Errorf("loop %s must observe, verify, and produce evidence", loop.ID)
	}
	if loop.ExpiresAfterHours <= 0 || len(loop.FailsWhen) == 0 {
		return fmt.Errorf("loop %s must declare expiry and fail conditions", loop.ID)
	}
	if loop.Kind == kindHarness && len(loop.PromotesTo) == 0 {
		return fmt.Errorf("harness loop %s must promote failures to candidates", loop.ID)
	}
	return verifyLoopSchedule(root, m, loop)
}
