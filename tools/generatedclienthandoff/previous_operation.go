package main

import "regexp"

func parsePreviousOperation(match []string) operationRow {
	op := operationRow{
		GeneratedPath: tsUnescape(match[1]),
		OperationID:   tsUnescape(match[2]),
		Method:        tsUnescape(match[3]),
		Path:          tsUnescape(match[4]),
		Deprecated:    match[5] == "true",
	}
	extra := match[6]
	if lifecycle := previousLifecyclePattern.FindStringSubmatch(extra); len(lifecycle) == 2 {
		op.Lifecycle = tsUnescape(lifecycle[1])
	}
	if replacement := previousReplacementPattern.FindStringSubmatch(extra); len(replacement) == 2 {
		op.Replacement = tsUnescape(replacement[1])
	}
	if horizon := previousRemovalHorizonPattern.FindStringSubmatch(extra); len(horizon) == 2 {
		op.RemovalHorizon = tsUnescape(horizon[1])
	}
	return op
}

var (
	previousLifecyclePattern      = regexp.MustCompile(`lifecycle:\s*'([^']*)'`)
	previousReplacementPattern    = regexp.MustCompile(`replacement:\s*'([^']*)'`)
	previousRemovalHorizonPattern = regexp.MustCompile(`removalHorizon:\s*'([^']*)'`)
)
