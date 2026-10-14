package main

import (
	"strings"
	"testing"
)

func verifyFigmaAPIGeneratedAnnotations(t *testing.T, annotations []figmaSourceAPIGeneratedAnnotation, docText string, surface figmaGeneratedClientSurface) {
	t.Helper()
	if got, want := len(annotations), 2; got != want {
		t.Fatalf("mirrored api_generated_annotations = %d, want %d", got, want)
	}
	seen := map[string]bool{}
	for _, annotation := range annotations {
		if seen[annotation.NodeID] {
			t.Fatalf("duplicate mirrored API Generated annotation %q", annotation.NodeID)
		}
		seen[annotation.NodeID] = true
		verifyFigmaAPIGeneratedAnnotation(t, annotation, docText, surface)
	}
}

func verifyFigmaAPIGeneratedAnnotation(t *testing.T, annotation figmaSourceAPIGeneratedAnnotation, docText string, surface figmaGeneratedClientSurface) {
	t.Helper()
	verifyFigmaAPIGeneratedAnnotationIdentity(t, annotation, surface)
	sourcePaths, ok := surface.SourceGeneratedPaths[annotation.CoverageEntryNodeID]
	if !ok || !sourcePaths[annotation.CanonicalGeneratedPath] {
		t.Fatalf("mirrored API Generated annotation %q canonical path %q is not covered by source entry %q", annotation.NodeID, annotation.CanonicalGeneratedPath, annotation.CoverageEntryNodeID)
	}
	v2Path := verifyV2Counterpart(t, "mirrored API Generated annotation "+annotation.NodeID, annotation.CanonicalGeneratedPath, sourcePaths, surface)
	verifyFacadeGeneratedClientComments(t, "mirrored API Generated annotation "+annotation.NodeID, annotation.CanonicalGeneratedPath, annotation.FigmaGeneratedPath, surface)
	verifyFacadeGeneratedClientComments(t, "mirrored API Generated annotation "+annotation.NodeID, v2Path, "riido."+v2Path, surface)
	requireDocMentions(t, docText, "mirrored API Generated annotation", []string{
		annotation.NodeID,
		annotation.FigmaGeneratedPath,
		annotation.CanonicalGeneratedPath,
		annotation.CategoryLabel,
	})
	verifyDocGeneratedPath(t, docText, "mirrored API Generated annotation v2 counterpart", v2Path)
	verifyFigmaAPIGeneratedAnnotationStaleCopy(t, annotation, docText)
}

func verifyFigmaAPIGeneratedAnnotationIdentity(t *testing.T, annotation figmaSourceAPIGeneratedAnnotation, surface figmaGeneratedClientSurface) {
	t.Helper()
	if annotation.CategoryID != "700:0" || annotation.CategoryLabel != "API Generated" {
		t.Fatalf("mirrored API Generated annotation %q category drifted: %+v", annotation.NodeID, annotation)
	}
	if !strings.HasPrefix(annotation.FigmaGeneratedPath, "riido.") {
		t.Fatalf("mirrored API Generated annotation %q must preserve Figma facade path: %q", annotation.NodeID, annotation.FigmaGeneratedPath)
	}
	canonical := canonicalPathFromFigmaFacade(annotation.FigmaGeneratedPath)
	if annotation.CanonicalGeneratedPath != canonical {
		t.Fatalf("mirrored API Generated annotation %q canonical path = %q, want %q", annotation.NodeID, annotation.CanonicalGeneratedPath, canonical)
	}
	verifyKnownGeneratedPath(t, "mirrored API Generated annotation "+annotation.NodeID, annotation.CanonicalGeneratedPath, surface.GeneratedPaths)
}

func verifyFigmaAPIGeneratedAnnotationStaleCopy(t *testing.T, annotation figmaSourceAPIGeneratedAnnotation, docText string) {
	t.Helper()
	if !strings.Contains(annotation.FigmaLabel, "작업중") {
		return
	}
	if annotation.ResolutionStatus != "resolved_stale_handoff_copy" || !strings.Contains(annotation.Resolution, "stale") {
		t.Fatalf("mirrored API Generated annotation %q stale copy is not resolved: %+v", annotation.NodeID, annotation)
	}
	requireDocMentions(t, docText, "stale Figma handoff copy", []string{"상세내용은 작업중입니다"})
}
