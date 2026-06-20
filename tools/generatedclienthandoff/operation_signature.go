package main

import (
	"fmt"
	"strings"
)

func operationSignature(op operationRow) string {
	parts := []string{
		op.Method,
		op.Path,
		op.OperationID,
		fmt.Sprintf("%t", op.Deprecated),
		op.Lifecycle,
		op.Replacement,
		op.RemovalHorizon,
	}
	return strings.Join(parts, "\x00")
}

func describeOperation(op operationRow) string {
	return fmt.Sprintf(
		"`%s` -> `%s %s` (`%s`, lifecycle: `%s`)",
		op.GeneratedPath,
		op.Method,
		op.Path,
		op.OperationID,
		lifecycleLabel(op),
	)
}

func describeChangedOperation(before, after operationRow) string {
	var parts []string
	parts = appendChangedHTTP(parts, before, after)
	parts = appendChangedIdentity(parts, before, after)
	parts = appendChangedLifecycle(parts, before, after)
	return fmt.Sprintf("`%s`: %s", after.GeneratedPath, strings.Join(parts, ", "))
}
