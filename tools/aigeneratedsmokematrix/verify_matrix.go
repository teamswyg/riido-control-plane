package main

import (
	"fmt"
	"strings"
)

func verifyMatrix(m manifest, ops map[string]generatedOperation, matrix smokeMatrix) error {
	if matrix.SchemaVersion != m.SmokeSchemaVersion {
		return fmt.Errorf("smoke schema = %q", matrix.SchemaVersion)
	}
	entries := map[string]smokeEntry{}
	previous := ""
	for _, entry := range matrix.Entries {
		if err := verifyEntry(m, ops, entry, previous); err != nil {
			return err
		}
		previous = entry.GeneratedPath + " " + entry.Method
		entries[entry.GeneratedPath] = entry
	}
	for path := range ops {
		if _, ok := entries[path]; !ok {
			return fmt.Errorf("OpenAPI generated path missing from smoke matrix: %s", path)
		}
	}
	return nil
}

func verifyEntry(m manifest, ops map[string]generatedOperation, entry smokeEntry, previous string) error {
	sortKey := entry.GeneratedPath + " " + entry.Method
	if previous != "" && sortKey < previous {
		return fmt.Errorf("smoke matrix is not sorted at %s", sortKey)
	}
	op, ok := ops[entry.GeneratedPath]
	if !ok {
		return fmt.Errorf("smoke matrix has unknown generated path %s", entry.GeneratedPath)
	}
	if entry.Method != op.Method || entry.Path != op.Path {
		return fmt.Errorf("smoke matrix drift for %s", entry.GeneratedPath)
	}
	return verifyEvidenceTests(m, entry)
}

func wantsV2(entry smokeEntry) bool { return strings.HasPrefix(entry.GeneratedPath, "v2.") }
