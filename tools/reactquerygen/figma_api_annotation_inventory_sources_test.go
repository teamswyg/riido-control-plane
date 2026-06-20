package main

import "testing"

func verifyFigmaAPIGeneratedInventoryGroupSources(t *testing.T, group figmaSourceAPIGeneratedAnnotationGroup, surface figmaGeneratedClientSurface) string {
	t.Helper()
	v2Path := "v2." + group.CanonicalGeneratedPath
	if _, ok := surface.GeneratedPaths[v2Path]; !ok {
		t.Fatalf("mirrored API Generated inventory group %q must keep v2 generated path counterpart %q", group.CanonicalGeneratedPath, v2Path)
	}
	annotationCount := 0
	for _, source := range group.Sources {
		verifyFigmaAPIGeneratedInventorySource(t, group, source, v2Path, surface)
		annotationCount += len(source.NodeIDs)
	}
	if group.AnnotationCount != annotationCount {
		t.Fatalf("mirrored API Generated inventory group %q annotation_count = %d, want node count %d", group.CanonicalGeneratedPath, group.AnnotationCount, annotationCount)
	}
	return v2Path
}

func verifyFigmaAPIGeneratedInventorySource(t *testing.T, group figmaSourceAPIGeneratedAnnotationGroup, source figmaSourceAPIGeneratedAnnotationSource, v2Path string, surface figmaGeneratedClientSurface) {
	t.Helper()
	if source.PageID == "" || source.TopLevelNodeID == "" || source.CoverageEntryNodeID == "" {
		t.Fatalf("mirrored API Generated inventory group %q has invalid source: %+v", group.CanonicalGeneratedPath, source)
	}
	sourcePaths, ok := surface.SourceGeneratedPaths[source.CoverageEntryNodeID]
	if !ok || !sourcePaths[group.CanonicalGeneratedPath] {
		t.Fatalf("mirrored API Generated inventory group %q canonical path is not covered by source entry %q", group.CanonicalGeneratedPath, source.CoverageEntryNodeID)
	}
	if !sourcePaths[v2Path] {
		t.Fatalf("mirrored API Generated inventory group %q v2 path %q is not covered by source entry %q", group.CanonicalGeneratedPath, v2Path, source.CoverageEntryNodeID)
	}
	if len(source.NodeIDs) == 0 {
		t.Fatalf("mirrored API Generated inventory group %q source %q must list node ids", group.CanonicalGeneratedPath, source.TopLevelNodeID)
	}
}
