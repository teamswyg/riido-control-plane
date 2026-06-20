package main

import "strings"

func hasLifecycleNote(op operationRow) bool {
	return op.Deprecated ||
		strings.TrimSpace(op.Lifecycle) != "" ||
		strings.TrimSpace(op.Replacement) != "" ||
		strings.TrimSpace(op.RemovalHorizon) != ""
}

func lifecycleAttrs(op operationRow) []string {
	var attrs []string
	if op.Deprecated {
		attrs = append(attrs, "deprecated")
	}
	if strings.TrimSpace(op.Lifecycle) != "" {
		attrs = append(attrs, "lifecycle="+op.Lifecycle)
	}
	if strings.TrimSpace(op.Replacement) != "" {
		attrs = append(attrs, "replacement="+op.Replacement)
	}
	if strings.TrimSpace(op.RemovalHorizon) != "" {
		attrs = append(attrs, "removal_horizon="+op.RemovalHorizon)
	}
	return attrs
}
