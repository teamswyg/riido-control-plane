package main

import (
	"strings"
	"testing"
)

func verifyMirroredFigmaInspectionMethod(t *testing.T, method figmaCoverageInspectionMethod, docText string) {
	t.Helper()
	if method.ID != "figma-plugin-api-page-registry.v1" {
		t.Fatalf("source coverage inspection method id = %q", method.ID)
	}
	if method.PageRegistryExpression != "figma.root.children" {
		t.Fatalf("source coverage page registry expression = %q", method.PageRegistryExpression)
	}
	if method.TopLevelChildCountExpression != "await figma.setCurrentPageAsync(page); page.children.length" {
		t.Fatalf("source coverage top-level child count expression = %q", method.TopLevelChildCountExpression)
	}
	rule := strings.ToLower(method.Rule)
	requireDocMentions(t, rule, "inspection rule", []string{
		"supporting evidence",
		"must not redefine page-level child counts",
		"lazy/unloaded",
	})
	requireDocMentions(t, docText, "mirrored inspection method", []string{
		"figma.root.children",
		"await figma.setCurrentPageAsync(page)",
		"page.children.length",
		"supporting evidence only",
		"lazy/unloaded",
	})
}
