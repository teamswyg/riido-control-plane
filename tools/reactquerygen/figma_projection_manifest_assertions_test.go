package main

import (
	"strings"
	"testing"
)

func verifyFigmaProjectionManifestHeader(t *testing.T, manifest figmaProjectionManifest, sourceCoverage figmaSourceCoverageManifest, docPath string) {
	t.Helper()
	if manifest.SchemaVersion != "riido-control-plane-figma-ai-agent-projection.v1" {
		t.Fatalf("schema_version = %q", manifest.SchemaVersion)
	}
	if manifest.ID != "figma-ai-agent-control-plane-generated-client-projection" {
		t.Fatalf("id = %q", manifest.ID)
	}
	if manifest.RiidoTask != "RIID-4810" {
		t.Fatalf("riido_task = %q", manifest.RiidoTask)
	}
	if manifest.EvidenceTool != "tools/figmaprojection" {
		t.Fatalf("evidence_tool = %q", manifest.EvidenceTool)
	}
	verifyFigmaProjectionSourceManifest(t, manifest.SourceContractsManifest)
	verifyFigmaProjectionSourceCoverageMirror(t, manifest, sourceCoverage, docPath)
}

func verifyFigmaProjectionSourceManifest(t *testing.T, source figmaProjectionSourceManifest) {
	t.Helper()
	if source.Repo != "riido-contracts" ||
		source.Path != "docs/30-architecture/figma-ai-agent-coverage.riido.json" ||
		source.SchemaVersion != "riido-figma-ai-agent-coverage.v1" ||
		source.ID != "figma-v1-22-ai-agent-ui-coverage" {
		t.Fatalf("source contracts manifest = %+v", source)
	}
}

func verifyFigmaProjectionSourceCoverageMirror(t *testing.T, manifest figmaProjectionManifest, sourceCoverage figmaSourceCoverageManifest, docPath string) {
	t.Helper()
	if sourceCoverage.SchemaVersion != manifest.SourceContractsManifest.SchemaVersion ||
		sourceCoverage.ID != manifest.SourceContractsManifest.ID {
		t.Fatalf("source coverage mirror = %s/%s, want %s/%s", sourceCoverage.SchemaVersion, sourceCoverage.ID, manifest.SourceContractsManifest.SchemaVersion, manifest.SourceContractsManifest.ID)
	}
	verifySourceContractsManifestProvenance(t, sourceCoverage.StabilizedBy, manifest.SourceContractsManifest.StabilizedBy, docPath)
}

func verifyFigmaProjectionSourceMirrors(t *testing.T, manifest figmaProjectionManifest, sourceCoverage figmaSourceCoverageManifest, docText string) {
	t.Helper()
	if got, want := len(sourceCoverage.ExpectedPages), 3; got != want {
		t.Fatalf("source coverage expected_pages = %d, want %d", got, want)
	}
	if got, want := len(sourceCoverage.NonUITopLevelNodes), 12; got != want {
		t.Fatalf("source coverage non_ui_top_level_nodes = %d, want %d", got, want)
	}
	verifyMirroredNonUITopLevelInventory(t, sourceCoverage)
	if strings.TrimSpace(manifest.ProjectionPolicy.TopDown) == "" || strings.TrimSpace(manifest.ProjectionPolicy.BottomUp) == "" {
		t.Fatalf("projection policy must include both directions: %+v", manifest.ProjectionPolicy)
	}
	if strings.TrimSpace(manifest.Loop.Observation) == "" || strings.TrimSpace(manifest.Loop.Retrospective) == "" {
		t.Fatalf("projection loop must be present: %+v", manifest.Loop)
	}
	verifyMirroredFigmaInspectionMethod(t, sourceCoverage.InspectionMethod, docText)
	verifyMirroredFigmaSupportingToolLimitations(t, manifest.MirroredSupportingToolLimitations, sourceCoverage.SupportingToolLimitations, manifest.SourceContractsManifest.StabilizedBy, docText)
	verifyMirroredFigmaAPIGeneratedAnnotationContentPolicy(t, sourceCoverage.APIGeneratedAnnotationContentPolicy, docText)
}
