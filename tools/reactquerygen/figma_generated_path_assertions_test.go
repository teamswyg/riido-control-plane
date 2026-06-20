package main

import (
	"strings"
	"testing"
)

func verifyKnownGeneratedPath(t *testing.T, owner, path string, generatedPaths map[string]string) {
	t.Helper()
	if route, ok := generatedPaths[path]; !ok {
		t.Fatalf("%s requires unknown generated path %q", owner, path)
	} else if strings.TrimSpace(route) == "" {
		t.Fatalf("%s generated path %q has empty route", owner, path)
	}
}

func verifyCoreReactGeneratedPathComments(t *testing.T, owner, path string, surface figmaGeneratedClientSurface) {
	t.Helper()
	requiredComment := "계약 generated path: `" + path + "`"
	if !strings.Contains(surface.Core, requiredComment) {
		t.Fatalf("core generated client missing %q for %s", requiredComment, owner)
	}
	if !strings.Contains(surface.React, requiredComment) {
		t.Fatalf("react generated client missing %q for %s", requiredComment, owner)
	}
}

func verifyFacadeGeneratedClientComments(t *testing.T, owner, canonical, facade string, surface figmaGeneratedClientSurface) {
	t.Helper()
	verifyGeneratedBodyComments(t, owner, "core", surface.Core, canonical, facade)
	verifyGeneratedBodyComments(t, owner, "react", surface.React, canonical, facade)
}

func verifyGeneratedBodyComments(t *testing.T, owner, name, body, canonical, facade string) {
	t.Helper()
	canonicalComment := "계약 generated path: `" + canonical + "`"
	accessComment := "접근 예시: `" + facade + "`"
	if !strings.Contains(body, canonicalComment) {
		t.Fatalf("%s generated client missing %q for %s", name, canonicalComment, owner)
	}
	if !strings.Contains(body, accessComment) {
		t.Fatalf("%s generated client missing %q for %s", name, accessComment, owner)
	}
}

func verifyV2Counterpart(t *testing.T, owner, canonical string, sourcePaths map[string]bool, surface figmaGeneratedClientSurface) string {
	t.Helper()
	v2Path := "v2." + canonical
	if _, ok := surface.GeneratedPaths[v2Path]; !ok {
		t.Fatalf("%s must keep v2 generated path counterpart %q", owner, v2Path)
	}
	if !sourcePaths[v2Path] {
		t.Fatalf("%s v2 path %q is not covered by source entry", owner, v2Path)
	}
	return v2Path
}

func verifyDocGeneratedPath(t *testing.T, docText, owner, path string) {
	t.Helper()
	if !docMentionsGeneratedPath(docText, path) {
		t.Fatalf("projection doc must mention %s generated path %q", owner, path)
	}
}
