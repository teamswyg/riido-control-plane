package main

import (
	"fmt"
	"strings"
)

func lifecycleLabel(op operationRow) string {
	if strings.TrimSpace(op.Lifecycle) != "" {
		return op.Lifecycle
	}
	if op.Deprecated {
		return "deprecated"
	}
	return "not_declared"
}

func operationLifecycleFields(op operationRow) string {
	var parts []string
	if strings.TrimSpace(op.Lifecycle) != "" {
		parts = append(parts, fmt.Sprintf("lifecycle: '%s'", ts(op.Lifecycle)))
	}
	if strings.TrimSpace(op.Replacement) != "" {
		parts = append(parts, fmt.Sprintf("replacement: '%s'", ts(op.Replacement)))
	}
	if strings.TrimSpace(op.RemovalHorizon) != "" {
		parts = append(parts, fmt.Sprintf("removalHorizon: '%s'", ts(op.RemovalHorizon)))
	}
	if len(parts) == 0 {
		return ""
	}
	return ", " + strings.Join(parts, ", ")
}

func lifecycleNotes(ops []operationRow) []string {
	var notes []string
	for _, op := range ops {
		if !hasLifecycleNote(op) {
			continue
		}
		notes = append(notes, fmt.Sprintf("`%s`: %s", op.GeneratedPath, strings.Join(lifecycleAttrs(op), ", ")))
	}
	return notes
}

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
