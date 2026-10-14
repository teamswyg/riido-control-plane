package main

import "testing"

func TestLineBudgetRatchetRejectsWorseMaxLines(t *testing.T) {
	err := verifyLineBudgetRatchet(lineBudgetResult{
		OverTarget:        2,
		MaxLines:          101,
		MaxFileLinesLimit: 100,
	})
	if err == nil {
		t.Fatal("expected ratchet failure")
	}
}

func TestLineBudgetRatchetAllowsFileCountIncreaseWhenSurfaceShrinks(t *testing.T) {
	err := verifyLineBudgetRatchet(lineBudgetResult{
		OverTarget:         2,
		MaxLines:           90,
		MaxFilesOverTarget: 1,
		MaxFileLinesLimit:  100,
	})
	if err != nil {
		t.Fatalf("expected file count increase to be informational: %v", err)
	}
}

func TestLineBudgetHotspotRatchetRejectsWorseSurface(t *testing.T) {
	err := verifyLineBudgetHotspotRatchets([]lineBudgetHotspotRatchet{
		{Path: "internal/example", FilesSlack: -1, MaxLinesSlack: 1, TotalOverSlack: -1},
	})
	if err == nil {
		t.Fatal("expected hotspot ratchet failure")
	}
}

func TestLineBudgetHotspotRatchetAllowsFileCountIncreaseWhenSurfaceShrinks(t *testing.T) {
	err := verifyLineBudgetHotspotRatchets([]lineBudgetHotspotRatchet{
		{Path: "internal/example", FilesSlack: -1, MaxLinesSlack: 1, TotalOverSlack: 1},
	})
	if err != nil {
		t.Fatalf("expected hotspot file count increase to be informational: %v", err)
	}
}
