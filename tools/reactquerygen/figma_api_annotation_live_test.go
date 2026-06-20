package main

import (
	"strings"
	"testing"
)

func verifyFigmaAPIGeneratedLiveInspection(t *testing.T, scan figmaSourceAPIGeneratedAnnotationLiveScan, docText string) {
	t.Helper()
	if scan.ObservedAt != "2026-06-03" || !strings.Contains(scan.Tool, "use_figma") || !strings.Contains(scan.Tool, "categoryId") {
		t.Fatalf("mirrored API Generated annotation live inspection provenance drifted: %+v", scan)
	}
	expected := figmaAPIGeneratedExpectedLiveCounts()
	if len(scan.PageCounts) != len(expected) {
		t.Fatalf("mirrored API Generated annotation page_counts = %d, want %d", len(scan.PageCounts), len(expected))
	}
	totalRiido, totalAPIGenerated := verifyFigmaAPIGeneratedLivePages(t, scan.PageCounts, expected, docText)
	if scan.TotalRiidoAnnotations != totalRiido || scan.TotalAPIGeneratedAnnotations != totalAPIGenerated {
		t.Fatalf("mirrored API Generated annotation live totals = riido:%d/api:%d, want riido:%d/api:%d", scan.TotalRiidoAnnotations, scan.TotalAPIGeneratedAnnotations, totalRiido, totalAPIGenerated)
	}
	if totalRiido != 90 || totalAPIGenerated != 90 {
		t.Fatalf("mirrored API Generated annotation live totals = riido:%d/api:%d, want 90/90", totalRiido, totalAPIGenerated)
	}
}

func verifyFigmaAPIGeneratedLivePages(t *testing.T, pages []figmaSourceAPIGeneratedAnnotationLivePageCounter, expected map[string]figmaSourceAPIGeneratedAnnotationLivePageCounter, docText string) (int, int) {
	t.Helper()
	var totalRiido, totalAPIGenerated int
	for _, page := range pages {
		want, ok := expected[page.PageID]
		if !ok {
			t.Fatalf("unexpected mirrored API Generated annotation live page count: %+v", page)
		}
		if page.PageName != want.PageName || page.RiidoAnnotationCount != want.RiidoAnnotationCount || page.APIGeneratedCount != want.APIGeneratedCount {
			t.Fatalf("mirrored API Generated annotation live page count for %s = %+v, want %+v", page.PageID, page, want)
		}
		if page.MissingOperationKind != 0 || page.MissingBackground != 0 {
			t.Fatalf("mirrored API Generated annotation live page count has missing content: %+v", page)
		}
		totalRiido += page.RiidoAnnotationCount
		totalAPIGenerated += page.APIGeneratedCount
		requireDocMentions(t, docText, "mirrored API Generated annotation live page count", []string{page.PageID, page.PageName})
	}
	return totalRiido, totalAPIGenerated
}
