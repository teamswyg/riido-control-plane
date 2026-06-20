package main

import "testing"

func TestFigmaAIAgentControlPlaneProjectionManifest(t *testing.T) {
	paths := newFigmaProjectionPaths()
	manifest := loadFigmaProjectionManifest(t, paths.manifest)
	sourceCoverage := loadFigmaSourceCoverageManifest(t, paths.sourceCoverage)

	verifyFigmaProjectionManifestHeader(t, manifest, sourceCoverage, paths.doc)
	docText := readFigmaProjectionDoc(t, paths.doc)
	verifyFigmaProjectionDocumentBoundaries(t, docText)
	verifyFigmaProjectionSourceMirrors(t, manifest, sourceCoverage, docText)
	verifyFigmaProjectionStaleBoundaries(t)

	spec, err := loadOpenAPI(paths.openAPI)
	if err != nil {
		t.Fatalf("loadOpenAPI: %v", err)
	}
	surface := loadFigmaGeneratedClientSurface(t, spec).WithSourceCoverage(sourceCoverage)
	verifyRuntimeEndpointLabelProjection(t, sourceCoverage, docText)
	verifyFigmaAPIGeneratedAnnotations(t, sourceCoverage.APIGeneratedAnnotations, docText, surface)
	verifyFigmaAPIGeneratedAnnotationInventory(t, sourceCoverage.APIGeneratedAnnotationInventory, docText, surface)
	verifyLegacyNonUIAbsorptions(t, manifest.LegacyNonUIAbsorptions, sourceCoverage, docText, surface)
	verifyNonUIPlanningAbsorptions(t, manifest.NonUIPlanningAbsorptions, sourceCoverage, docText, surface)
	verifyFigmaProjectionEntries(t, manifest.Entries, docText, surface)
}
