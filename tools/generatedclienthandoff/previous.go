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
	body := string(data)
	if strings.TrimSpace(body) == "" {
		return previousManifest{}, nil
	}
	return parsePreviousManifest(body), nil
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
	previousCommitPattern    = regexp.MustCompile(`sourceCommit:\s*'([^']*)'`)
	previousOperationPattern = regexp.MustCompile(
		`(?s)\{\s*generatedPath:\s*'([^']*)',\s*operationId:\s*'([^']*)',\s*method:\s*'([^']*)',\s*path:\s*'([^']*)',\s*deprecated:\s*(true|false)([^}]*)\}`,
	)
	previousLifecyclePattern      = regexp.MustCompile(`lifecycle:\s*'([^']*)'`)
	previousReplacementPattern    = regexp.MustCompile(`replacement:\s*'([^']*)'`)
	previousRemovalHorizonPattern = regexp.MustCompile(`removalHorizon:\s*'([^']*)'`)
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
