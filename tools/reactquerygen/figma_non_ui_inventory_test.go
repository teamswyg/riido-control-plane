package main

import (
	"strconv"
	"strings"
	"testing"
)

func verifyMirroredNonUITopLevelInventory(t *testing.T, sourceCoverage figmaSourceCoverageManifest) {
	t.Helper()
	pages := map[string]figmaSourceCoveragePage{}
	for _, page := range sourceCoverage.ExpectedPages {
		pages[page.NodeID] = page
	}
	if got, want := len(sourceCoverage.NonUITopLevelInventory), 2; got != want {
		t.Fatalf("source coverage non_ui_top_level_inventory pages = %d, want %d", got, want)
	}
	for _, inventory := range sourceCoverage.NonUITopLevelInventory {
		page, ok := pages[inventory.PageID]
		if !ok {
			t.Fatalf("non-UI inventory references unknown page %q", inventory.PageID)
		}
		if got, want := len(inventory.Nodes), page.ChildCount; got != want {
			if !sourceFigmaNonUIInventoryDriftDocumented(sourceCoverage.SupportingToolLimitations, inventory.PageID, got, want) {
				t.Fatalf("non-UI inventory page %q nodes = %d, want loaded child_count %d", inventory.PageID, got, want)
			}
		}
	}
	wireframe := pages["0:1"]
	if wireframe.ChildCount != 28 {
		t.Fatalf("Wireframe page loaded child_count = %d, want 28", wireframe.ChildCount)
	}
}

func sourceFigmaNonUIInventoryDriftDocumented(limitations []figmaSourceSupportingToolLimitation, pageID string, knownInventoryCount, childCount int) bool {
	for _, limitation := range limitations {
		if limitation.ID != "figma-onboarding-page-load-timeout.v1" || !hasString(limitation.AuthoritativeResult, pageID) {
			continue
		}
		expected := []string{
			"child_count=" + strconv.Itoa(childCount),
			"known_inventory_count=" + strconv.Itoa(knownInventoryCount),
			"unresolved_extra_top_level_node=" + strconv.Itoa(childCount-knownInventoryCount),
		}
		if hasAllStrings(limitation.AuthoritativeResult, expected) {
			return strings.Contains(strings.ToLower(limitation.Rule), "known_inventory_count may lag expected_pages.child_count")
		}
	}
	return false
}

func hasAllStrings(items, expected []string) bool {
	for _, item := range expected {
		if !hasString(items, item) {
			return false
		}
	}
	return true
}
