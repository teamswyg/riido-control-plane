package main

import "fmt"

func verifyOperations(m manifest) (int, error) {
	seen := map[string]bool{}
	for _, item := range m.OperationEvidence {
		if item.Route == "" || item.Intent == "" {
			return 0, required("operation route and intent")
		}
		for _, op := range item.StoreOperations {
			seen[op] = true
		}
	}
	for _, op := range requiredOperations {
		if !seen[op] {
			return 0, fmt.Errorf("missing operation evidence %q", op)
		}
	}
	return len(seen), nil
}

func verifySignals(m manifest) (int, error) {
	for _, signal := range requiredSignals {
		if !containsText(m.MeasurementSignals, signal) {
			return 0, fmt.Errorf("missing measurement signal %q", signal)
		}
	}
	return len(m.MeasurementSignals), nil
}
