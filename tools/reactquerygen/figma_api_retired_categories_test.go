package main

import (
	"strings"
	"testing"
)

func verifyMirroredFigmaAPIGeneratedRetiredCategories(t *testing.T, categories []figmaSourceAPIGeneratedAnnotationRetiredCategory, docText string) {
	t.Helper()
	if len(categories) != 1 {
		t.Fatalf("mirrored API Generated retired categories = %d, want 1", len(categories))
	}
	retired := categories[0]
	if retired.CategoryID != "39:0" || retired.CategoryLabel != "클라이언트 전달" {
		t.Fatalf("unexpected mirrored retired API Generated category: %+v", retired)
	}
	if retired.RetirementStatus != "unused_not_deleted" || retired.LiveUsageCount != 0 {
		t.Fatalf("mirrored retired API Generated category must stay unused_not_deleted with zero live usage: %+v", retired)
	}
	if retired.ObservedAt != "2026-06-03" || !strings.Contains(retired.ToolLimitation, "design owner") {
		t.Fatalf("mirrored retired API Generated category must record automation limitation: %+v", retired)
	}
	requireDocMentions(t, docText, "mirrored retired API Generated category", []string{
		retired.CategoryID,
		retired.CategoryLabel,
		"retired",
		"0",
	})
}
