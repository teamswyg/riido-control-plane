package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

func readPreviousManifest(path string) (previousManifest, error) {
	if strings.TrimSpace(path) == "" {
		return previousManifest{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return previousManifest{}, nil
		}
		return previousManifest{}, fmt.Errorf("read previous manifest: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return previousManifest{}, nil
	}
	return parsePreviousManifest(string(data)), nil
}

func parsePreviousManifest(body string) previousManifest {
	manifest := previousManifest{Available: true}
	if match := previousCommitPattern.FindStringSubmatch(body); len(match) == 2 {
		manifest.SourceCommit = tsUnescape(match[1])
	}
	for _, match := range previousOperationPattern.FindAllStringSubmatch(body, -1) {
		manifest.Operations = append(manifest.Operations, parsePreviousOperation(match))
	}
	sort.Slice(manifest.Operations, func(i, j int) bool {
		return manifest.Operations[i].GeneratedPath < manifest.Operations[j].GeneratedPath
	})
	return manifest
}

var (
	tsQuoted                 = `((?:\\'|[^'])*)`
	previousCommitPattern    = regexp.MustCompile(`sourceCommit:\s*'` + tsQuoted + `'`)
	previousOperationPattern = regexp.MustCompile(
		`(?s)\{\s*generatedPath:\s*'` + tsQuoted + `',\s*operationId:\s*'` +
			tsQuoted + `',\s*method:\s*'` + tsQuoted + `',\s*path:\s*'` +
			tsQuoted + `',\s*deprecated:\s*(true|false)([^}]*)\}`,
	)
	previousLifecyclePattern      = regexp.MustCompile(`lifecycle:\s*'` + tsQuoted + `'`)
	previousReplacementPattern    = regexp.MustCompile(`replacement:\s*'` + tsQuoted + `'`)
	previousRemovalHorizonPattern = regexp.MustCompile(`removalHorizon:\s*'` + tsQuoted + `'`)
)

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
