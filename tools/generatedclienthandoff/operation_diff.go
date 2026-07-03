package main

import (
	"fmt"
	"sort"
	"strings"
)

func operationDiff(previous, current []operationRow) (added, removed, changed []string) {
	previousByPath := map[string]operationRow{}
	currentByPath := map[string]operationRow{}
	for _, op := range previous {
		previousByPath[op.GeneratedPath] = op
	}
	for _, op := range current {
		currentByPath[op.GeneratedPath] = op
		if before, ok := previousByPath[op.GeneratedPath]; !ok {
			added = append(added, describeOperation(op))
		} else if operationSignature(before) != operationSignature(op) {
			changed = append(changed, describeChangedOperation(before, op))
		}
	}
	for _, op := range previous {
		if _, ok := currentByPath[op.GeneratedPath]; !ok {
			removed = append(removed, describeOperation(op))
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(changed)
	return added, removed, changed
}

func operationSignature(op operationRow) string {
	parts := []string{op.Method, op.Path, op.OperationID, fmt.Sprintf("%t", op.Deprecated)}
	parts = append(parts, op.Lifecycle, op.Replacement, op.RemovalHorizon)
	return strings.Join(parts, "\x00")
}

func describeOperation(op operationRow) string {
	return fmt.Sprintf("`%s` -> `%s %s` (`%s`, lifecycle: `%s`)", op.GeneratedPath, op.Method, op.Path, op.OperationID, lifecycleLabel(op))
}

func describeChangedOperation(before, after operationRow) string {
	var parts []string
	parts = appendChangedHTTP(parts, before, after)
	parts = appendChangedIdentity(parts, before, after)
	parts = appendChangedLifecycle(parts, before, after)
	return fmt.Sprintf("`%s`: %s", after.GeneratedPath, strings.Join(parts, ", "))
}
