package main

import "sort"

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
