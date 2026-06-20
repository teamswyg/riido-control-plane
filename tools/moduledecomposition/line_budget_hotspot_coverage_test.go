package main

import "testing"

func TestLineBudgetHotspotCoverageRejectsUntrackedHotspot(t *testing.T) {
	err := verifyLineBudgetHotspotCoverage([]lineBudgetHotspot{
		{Path: "tools/newhotspot", Files: 1, MaxLines: 100, TotalOver: 25},
	})
	if err == nil {
		t.Fatal("expected untracked hotspot failure")
	}
}

func TestUntrackedLineBudgetHotspotsIgnoresTrackedHotspots(t *testing.T) {
	got := untrackedLineBudgetHotspots(
		[]lineBudgetHotspot{{Path: "tools/known"}, {Path: "tools/new"}},
		[]lineBudgetHotspotLimit{{Path: "tools/known"}},
	)
	if len(got) != 1 || got[0].Path != "tools/new" {
		t.Fatalf("unexpected untracked hotspots: %+v", got)
	}
}
