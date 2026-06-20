package main

import (
	"strings"
	"testing"
)

func verifyFigmaProjectionEntries(t *testing.T, entries []figmaProjectionEntry, docText string, surface figmaGeneratedClientSurface) {
	t.Helper()
	if got, want := len(entries), 16; got != want {
		t.Fatalf("entries = %d, want %d", got, want)
	}
	seen := map[string]bool{}
	for _, entry := range entries {
		if seen[entry.NodeID] {
			t.Fatalf("duplicate node_id %q", entry.NodeID)
		}
		seen[entry.NodeID] = true
		requireDocMentions(t, docText, "projection entry", []string{entry.NodeID, entry.Name})
		verifyFigmaProjectionEntry(t, entry, surface)
	}
}

func verifyFigmaProjectionEntry(t *testing.T, entry figmaProjectionEntry, surface figmaGeneratedClientSurface) {
	t.Helper()
	if strings.TrimSpace(entry.NodeID) == "" || strings.TrimSpace(entry.Name) == "" {
		t.Fatalf("entry has empty node id or name: %+v", entry)
	}
	if strings.TrimSpace(entry.SourceCoverageStatus) == "" {
		t.Fatalf("entry %q source_coverage_status is required", entry.NodeID)
	}
	switch entry.ProjectionStatus {
	case "generated_client_covered":
		verifyGeneratedClientCoveredProjectionEntry(t, entry, surface)
	case "client_route_no_endpoint", "product_surface_no_endpoint", "planning_evidence", "non_decision_asset":
		verifyNonEndpointProjectionEntry(t, entry)
	default:
		t.Fatalf("entry %q has unknown projection_status %q", entry.NodeID, entry.ProjectionStatus)
	}
	for _, fragment := range entry.ForbiddenGeneratedPathFragments {
		if strings.Contains(surface.GeneratedHaystack, strings.ToLower(fragment)) {
			t.Fatalf("entry %q forbids generated path fragment %q, but generated surface contains it", entry.NodeID, fragment)
		}
	}
}

func verifyNonEndpointProjectionEntry(t *testing.T, entry figmaProjectionEntry) {
	t.Helper()
	if strings.TrimSpace(entry.NoEndpointReason) == "" {
		t.Fatalf("entry %q must explain why no endpoint is generated", entry.NodeID)
	}
	if len(entry.RequiredGeneratedPaths) != 0 {
		t.Fatalf("entry %q must not require generated paths for status %s", entry.NodeID, entry.ProjectionStatus)
	}
}
