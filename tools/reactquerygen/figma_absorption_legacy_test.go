package main

import (
	"strings"
	"testing"
)

func verifyLegacyNonUIAbsorptions(t *testing.T, absorptions []figmaProjectionLegacyAbsorption, sourceCoverage figmaSourceCoverageManifest, docText string, surface figmaGeneratedClientSurface) {
	t.Helper()
	if got, want := len(absorptions), 7; got != want {
		t.Fatalf("legacy_non_ui_absorptions = %d, want %d", got, want)
	}
	sourcePrimary := figmaCoverageEntriesByNode(sourceCoverage.Entries)
	sourceNonUI := figmaCoverageEntriesByNode(sourceCoverage.NonUITopLevelNodes)
	seen := map[string]bool{}
	for _, absorption := range absorptions {
		if seen[absorption.NodeID] {
			t.Fatalf("duplicate legacy absorption node_id %q", absorption.NodeID)
		}
		seen[absorption.NodeID] = true
		verifyLegacyAbsorption(t, absorption, sourcePrimary, sourceNonUI, docText, surface)
	}
}

func verifyLegacyAbsorption(t *testing.T, absorption figmaProjectionLegacyAbsorption, sourcePrimary, sourceNonUI map[string]figmaSourceCoverageEntry, docText string, surface figmaGeneratedClientSurface) {
	t.Helper()
	if absorption.ProjectionStatus != "absorbed_by_current_ui_generated_client" {
		t.Fatalf("legacy absorption %q projection_status = %q", absorption.NodeID, absorption.ProjectionStatus)
	}
	if strings.TrimSpace(absorption.LocalScope) == "" {
		t.Fatalf("legacy absorption %q must explain local_scope", absorption.NodeID)
	}
	if len(absorption.RequiredGeneratedPaths) == 0 {
		t.Fatalf("legacy absorption %q must require generated paths", absorption.NodeID)
	}
	source, ok := sourceNonUI[absorption.NodeID]
	if !ok {
		t.Fatalf("legacy absorption %q is missing from mirrored contracts non_ui_top_level_nodes", absorption.NodeID)
	}
	verifyLegacyAbsorptionSource(t, absorption, source, sourcePrimary)
	verifyLegacyAbsorptionGeneratedPaths(t, absorption, source, sourcePrimary[absorption.AbsorbedByTopLevelNodeID], surface)
	requireDocMentions(t, docText, "legacy absorption", []string{
		absorption.NodeID,
		absorption.Name,
		absorption.AbsorbedByTopLevelNodeID,
	})
}

func verifyLegacyAbsorptionSource(t *testing.T, absorption figmaProjectionLegacyAbsorption, source figmaSourceCoverageEntry, sourcePrimary map[string]figmaSourceCoverageEntry) {
	t.Helper()
	if source.PageID != "0:1" || source.CoverageStatus != "covered" || source.EvidenceKind != "figma_legacy_wireframe_section" {
		t.Fatalf("legacy absorption %q source coverage is not a covered legacy Wireframe section: %+v", absorption.NodeID, source)
	}
	if absorption.SourceCoverageStatus != source.CoverageStatus || source.Name != absorption.Name {
		t.Fatalf("legacy absorption %q source mirror drifted: absorption=%+v source=%+v", absorption.NodeID, absorption, source)
	}
	if source.AbsorbedByTopLevelNodeID != absorption.AbsorbedByTopLevelNodeID {
		t.Fatalf("legacy absorption %q absorbed_by_top_level_node_id = %q, source = %q", absorption.NodeID, absorption.AbsorbedByTopLevelNodeID, source.AbsorbedByTopLevelNodeID)
	}
	if _, ok := sourcePrimary[absorption.AbsorbedByTopLevelNodeID]; !ok {
		t.Fatalf("legacy absorption %q references missing current UI source entry %q", absorption.NodeID, absorption.AbsorbedByTopLevelNodeID)
	}
}

func verifyLegacyAbsorptionGeneratedPaths(t *testing.T, absorption figmaProjectionLegacyAbsorption, source, absorbed figmaSourceCoverageEntry, surface figmaGeneratedClientSurface) {
	t.Helper()
	for _, path := range absorption.RequiredGeneratedPaths {
		if !hasString(source.GeneratedPaths, path) || !hasString(absorbed.GeneratedPaths, path) {
			t.Fatalf("legacy absorption %q requires generated path %q absent from mirrored source coverage", absorption.NodeID, path)
		}
		verifyKnownGeneratedPath(t, "legacy absorption "+absorption.NodeID, path, surface.GeneratedPaths)
		verifyCoreReactGeneratedPathComments(t, "legacy absorption "+absorption.NodeID, path, surface)
	}
}
