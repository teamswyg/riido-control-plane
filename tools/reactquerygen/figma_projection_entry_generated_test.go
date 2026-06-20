package main

import (
	"strings"
	"testing"
)

func verifyGeneratedClientCoveredProjectionEntry(t *testing.T, entry figmaProjectionEntry, surface figmaGeneratedClientSurface) {
	t.Helper()
	if strings.TrimSpace(entry.LocalScope) == "" {
		t.Fatalf("entry %q local_scope is required", entry.NodeID)
	}
	if len(entry.RequiredGeneratedPaths) == 0 {
		t.Fatalf("entry %q must require generated paths", entry.NodeID)
	}
	sourcePaths, ok := surface.SourceGeneratedPaths[entry.NodeID]
	if !ok {
		t.Fatalf("entry %q is missing from mirrored contracts Figma coverage", entry.NodeID)
	}
	for _, path := range entry.RequiredGeneratedPaths {
		verifyProjectionEntryGeneratedPath(t, entry, path, sourcePaths, surface)
	}
}

func verifyProjectionEntryGeneratedPath(t *testing.T, entry figmaProjectionEntry, path string, sourcePaths map[string]bool, surface figmaGeneratedClientSurface) {
	t.Helper()
	if !sourcePaths[path] {
		t.Fatalf("entry %q requires generated path %q that is absent from mirrored contracts Figma coverage", entry.NodeID, path)
	}
	verifyKnownGeneratedPath(t, "entry "+entry.NodeID, path, surface.GeneratedPaths)
	verifyCoreReactGeneratedPathComments(t, "entry "+entry.NodeID, path, surface)
}
