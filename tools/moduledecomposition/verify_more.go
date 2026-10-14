package main

import (
	"fmt"
	"strings"
)

func verifyPackageParity(actual []string, entries []packageEntry) error {
	seen := map[string]bool{}
	for _, entry := range entries {
		if entry.Path == "" || entry.Kind == "" || entry.Role == "" || entry.MustNotOwn == "" {
			return fmt.Errorf("package entry must include path, kind, role, and must_not_own")
		}
		if seen[entry.Path] {
			return fmt.Errorf("duplicate package entry %s", entry.Path)
		}
		seen[entry.Path] = true
	}
	for _, path := range actual {
		if !seen[path] {
			return fmt.Errorf("package %s missing from module manifest", path)
		}
	}
	if len(actual) != len(entries) {
		return fmt.Errorf("package count mismatch: actual=%d manifest=%d", len(actual), len(entries))
	}
	return nil
}

func summarizePackages(entries []packageEntry) verifyResult {
	result := verifyResult{PackageCount: len(entries)}
	for _, entry := range entries {
		switch {
		case strings.HasPrefix(entry.Path, "cmd/"):
			result.RuntimePackages++
		case strings.HasPrefix(entry.Path, "internal/"):
			result.InternalPackages++
		case strings.HasPrefix(entry.Path, "tools/"):
			result.ToolPackages++
		}
	}
	return result
}
