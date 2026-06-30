package main

import "fmt"

func verifyMeasurements(checkID string, measurements []measurement) error {
	if len(measurements) == 0 {
		return fmt.Errorf("readiness check %s must bind at least one measurement", checkID)
	}
	seen := map[string]bool{}
	for _, m := range measurements {
		if m.ID == "" || m.Kind == "" || m.Signal == "" {
			return fmt.Errorf("readiness check %s has incomplete measurement", checkID)
		}
		if seen[m.ID] {
			return fmt.Errorf("readiness check %s has duplicate measurement %s", checkID, m.ID)
		}
		seen[m.ID] = true
		if !allowedMeasurementKind(m.Kind) {
			return fmt.Errorf("readiness check %s has unknown measurement kind %s", checkID, m.Kind)
		}
	}
	return nil
}

func allowedMeasurementKind(kind string) bool {
	switch kind {
	case "artifact", "infra", "manual", "metric", "notion", "profile", "screenshot", "test", "trace", "workflow":
		return true
	default:
		return false
	}
}
