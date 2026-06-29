package main

import "testing"

func TestLoopCoverageDimensionRegistryHasExecutableHandlers(t *testing.T) {
	seen := map[string]bool{}
	for _, dim := range loopCoverageDimensions {
		if dim.id == "" || dim.loopField == "" || dim.claimField == "" ||
			dim.loopTokenLabel == "" || dim.claimTokenLabel == "" || dim.loopTokens == nil ||
			dim.claimTokens == nil {
			t.Fatalf("incomplete loop coverage dimension: %+v", dim)
		}
		if seen[dim.id] {
			t.Fatalf("duplicate loop coverage dimension: %s", dim.id)
		}
		seen[dim.id] = true
	}
}

func TestLoopCoverageDimensionRegistryCoversKnownAxes(t *testing.T) {
	want := []string{"observes", "verifies", "fails_when", "evidence"}
	for _, id := range want {
		if dim := loopCoverageDimensionByID(id); dim.id == "" {
			t.Fatalf("missing loop coverage dimension %s", id)
		}
	}
}
