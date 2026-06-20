package main

import "fmt"

func appendChangedHTTP(parts []string, before, after operationRow) []string {
	if before.Method != after.Method || before.Path != after.Path {
		return append(parts, fmt.Sprintf("HTTP `%s %s` -> `%s %s`", before.Method, before.Path, after.Method, after.Path))
	}
	return parts
}

func appendChangedIdentity(parts []string, before, after operationRow) []string {
	if before.OperationID != after.OperationID {
		parts = append(parts, fmt.Sprintf("operationId `%s` -> `%s`", before.OperationID, after.OperationID))
	}
	if before.Deprecated != after.Deprecated {
		parts = append(parts, fmt.Sprintf("deprecated `%t` -> `%t`", before.Deprecated, after.Deprecated))
	}
	return parts
}

func appendChangedLifecycle(parts []string, before, after operationRow) []string {
	if lifecycleLabel(before) != lifecycleLabel(after) {
		parts = append(parts, fmt.Sprintf("lifecycle `%s` -> `%s`", lifecycleLabel(before), lifecycleLabel(after)))
	}
	if before.Replacement != after.Replacement {
		parts = append(parts, fmt.Sprintf("replacement `%s` -> `%s`", before.Replacement, after.Replacement))
	}
	if before.RemovalHorizon != after.RemovalHorizon {
		parts = append(parts, fmt.Sprintf("removal horizon `%s` -> `%s`", before.RemovalHorizon, after.RemovalHorizon))
	}
	return parts
}
