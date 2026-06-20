package main

import (
	"strings"
	"testing"
)

func verifyFigmaAPIGeneratedAnnotationInventory(t *testing.T, inventory []figmaSourceAPIGeneratedAnnotationGroup, docText string, surface figmaGeneratedClientSurface) {
	t.Helper()
	if got, want := len(inventory), 20; got != want {
		t.Fatalf("mirrored api_generated_annotation_inventory = %d, want %d", got, want)
	}
	seenPath := map[string]bool{}
	totalAnnotations := 0
	for _, group := range inventory {
		verifyFigmaAPIGeneratedInventoryGroupIdentity(t, group, seenPath, surface)
		v2Path := verifyFigmaAPIGeneratedInventoryGroupSources(t, group, surface)
		verifyFigmaAPIGeneratedInventoryGeneratedClients(t, group, v2Path, surface)
		requireDocMentions(t, docText, "mirrored API Generated inventory", []string{
			group.UIArea,
			group.FigmaGeneratedPath,
			group.CanonicalGeneratedPath,
			group.OperationKind,
			group.Background,
		})
		verifyDocGeneratedPath(t, docText, "mirrored API Generated inventory v2 counterpart", v2Path)
		totalAnnotations += group.AnnotationCount
	}
	if got, want := totalAnnotations, 90; got != want {
		t.Fatalf("mirrored API Generated inventory node annotations = %d, want %d", got, want)
	}
}

func verifyFigmaAPIGeneratedInventoryGroupIdentity(t *testing.T, group figmaSourceAPIGeneratedAnnotationGroup, seenPath map[string]bool, surface figmaGeneratedClientSurface) {
	t.Helper()
	if strings.TrimSpace(group.UIArea) == "" {
		t.Fatalf("mirrored API Generated inventory group has empty ui_area: %+v", group)
	}
	if group.CategoryID != "700:0" || group.CategoryLabel != "API Generated" {
		t.Fatalf("mirrored API Generated inventory group %q category drifted: %+v", group.FigmaGeneratedPath, group)
	}
	if !strings.HasPrefix(group.FigmaGeneratedPath, "riido.") {
		t.Fatalf("mirrored API Generated inventory group must preserve Figma facade path: %q", group.FigmaGeneratedPath)
	}
	canonical := canonicalPathFromFigmaFacade(group.FigmaGeneratedPath)
	if group.CanonicalGeneratedPath != canonical {
		t.Fatalf("mirrored API Generated inventory group %q canonical path = %q, want %q", group.FigmaGeneratedPath, group.CanonicalGeneratedPath, canonical)
	}
	if seenPath[group.CanonicalGeneratedPath] {
		t.Fatalf("duplicate mirrored API Generated inventory generated path %q", group.CanonicalGeneratedPath)
	}
	seenPath[group.CanonicalGeneratedPath] = true
	verifyKnownGeneratedPath(t, "mirrored API Generated inventory "+group.CanonicalGeneratedPath, group.CanonicalGeneratedPath, surface.GeneratedPaths)
	if !figmaAPIGeneratedAllowedOperationKinds()[group.OperationKind] {
		t.Fatalf("mirrored API Generated inventory group %q operation_kind = %q", group.CanonicalGeneratedPath, group.OperationKind)
	}
	if strings.TrimSpace(group.Background) == "" {
		t.Fatalf("mirrored API Generated inventory group %q must preserve Korean background text", group.CanonicalGeneratedPath)
	}
}

func figmaAPIGeneratedAllowedOperationKinds() map[string]bool {
	return map[string]bool{"Query": true, "Mutation": true, "SSE Stream": true}
}
