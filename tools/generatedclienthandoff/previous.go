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
)
