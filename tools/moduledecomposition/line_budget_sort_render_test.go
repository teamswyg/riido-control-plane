package main

import (
	"strings"
	"testing"
)

func TestCompareLineBudgetHotspotsOrdersDeterministically(t *testing.T) {
	t.Parallel()
	base := lineBudgetHotspot{Path: "b", Files: 2, MaxLines: 10, TotalOver: 3}
	moreFiles := lineBudgetHotspot{Path: "a", Files: 3, MaxLines: 1, TotalOver: 1}
	if compareLineBudgetHotspots(base, moreFiles) <= 0 {
		t.Fatal("more files should sort first")
	}
	moreOver := lineBudgetHotspot{Path: "a", Files: 2, MaxLines: 1, TotalOver: 4}
	if compareLineBudgetHotspots(base, moreOver) <= 0 {
		t.Fatal("higher total over target should sort first")
	}
	moreLines := lineBudgetHotspot{Path: "a", Files: 2, MaxLines: 11, TotalOver: 3}
	if compareLineBudgetHotspots(base, moreLines) <= 0 {
		t.Fatal("higher max lines should sort first")
	}
	if compareLineBudgetHotspots(base, lineBudgetHotspot{Path: "a", Files: 2, MaxLines: 10, TotalOver: 3}) <= 0 {
		t.Fatal("path lexical order should break ties")
	}
}

func TestTrimLineBudgetHotspotsAndRenderUntracked(t *testing.T) {
	t.Parallel()
	hotspots := []lineBudgetHotspot{{Path: "a"}, {Path: "b"}}
	if got := trimLineBudgetHotspots(hotspots, 1); len(got) != 1 || got[0].Path != "a" {
		t.Fatalf("trimmed hotspots = %+v", got)
	}
	if got := trimLineBudgetHotspots(hotspots, 0); len(got) != 2 {
		t.Fatalf("limit zero should keep all hotspots: %+v", got)
	}
	var b strings.Builder
	renderLineBudgetUntrackedHotspots(&b, []lineBudgetHotspot{
		{Path: "internal/big", Files: 2, MaxLines: 90, TotalOver: 30},
	})
	out := b.String()
	if !strings.Contains(out, "internal/big") || !strings.Contains(out, "90") {
		t.Fatalf("missing hotspot table content: %s", out)
	}
}
