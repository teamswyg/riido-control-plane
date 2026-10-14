package main

import (
	"strings"
	"testing"
)

func verifyNonUIPlanningAbsorptions(t *testing.T, absorptions []figmaProjectionPlanningAbsorption, sourceCoverage figmaSourceCoverageManifest, docText string, surface figmaGeneratedClientSurface) {
	t.Helper()
	if got, want := len(absorptions), 1; got != want {
		t.Fatalf("non_ui_planning_absorptions = %d, want %d", got, want)
	}
	sourceNonUI := figmaCoverageEntriesByNode(sourceCoverage.NonUITopLevelNodes)
	seen := map[string]bool{}
	for _, absorption := range absorptions {
		if seen[absorption.NodeID] {
			t.Fatalf("duplicate planning absorption node_id %q", absorption.NodeID)
		}
		seen[absorption.NodeID] = true
		verifyPlanningAbsorption(t, absorption, sourceNonUI, docText, surface)
	}
}

func verifyPlanningAbsorption(t *testing.T, absorption figmaProjectionPlanningAbsorption, sourceNonUI map[string]figmaSourceCoverageEntry, docText string, surface figmaGeneratedClientSurface) {
	t.Helper()
	if absorption.ProjectionStatus != "absorbed_by_onboarding_generated_client" {
		t.Fatalf("planning absorption %q projection_status = %q", absorption.NodeID, absorption.ProjectionStatus)
	}
	if strings.TrimSpace(absorption.LocalScope) == "" || strings.TrimSpace(absorption.NoNewEndpointReason) == "" {
		t.Fatalf("planning absorption %q must explain local scope and no_new_endpoint_reason: %+v", absorption.NodeID, absorption)
	}
	source, ok := sourceNonUI[absorption.NodeID]
	if !ok {
		t.Fatalf("planning absorption %q is missing from mirrored contracts non_ui_top_level_nodes", absorption.NodeID)
	}
	verifyPlanningAbsorptionSource(t, absorption, source)
	for _, path := range absorption.RequiredGeneratedPaths {
		if !hasString(source.GeneratedPaths, path) {
			t.Fatalf("planning absorption %q requires generated path %q absent from mirrored non-UI source coverage", absorption.NodeID, path)
		}
		verifyKnownGeneratedPath(t, "planning absorption "+absorption.NodeID, path, surface.GeneratedPaths)
		verifyCoreReactGeneratedPathComments(t, "planning absorption "+absorption.NodeID, path, surface)
	}
	requireDocMentions(t, docText, "planning absorption boundary", []string{
		absorption.NodeID,
		absorption.Name,
		"client-local",
		"workspace-less create",
	})
}

func verifyPlanningAbsorptionSource(t *testing.T, absorption figmaProjectionPlanningAbsorption, source figmaSourceCoverageEntry) {
	t.Helper()
	if source.PageID != "42:3014" || source.CoverageStatus != "covered" || source.EvidenceKind != "figma_planning_section" {
		t.Fatalf("planning absorption %q source coverage is not a covered planning section: %+v", absorption.NodeID, source)
	}
	if source.Name != absorption.Name || absorption.SourceCoverageStatus != source.CoverageStatus {
		t.Fatalf("planning absorption %q source mirror drifted: absorption=%+v source=%+v", absorption.NodeID, absorption, source)
	}
}
