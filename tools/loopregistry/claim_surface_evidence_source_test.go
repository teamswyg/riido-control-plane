package main

import "testing"

func TestClaimSurfaceEvidenceIncludesEvidenceSourceCoverage(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll(root, m, hashes)
	if err != nil {
		t.Fatalf("verifyAll: %v", err)
	}
	got := newEvidence(m, result, nil)
	claimLoops := claimLoopIndex(m.Claims)
	covered := map[string]map[string]bool{}
	for _, surface := range got.ClaimSurfaces {
		loopID := claimLoops[surface.ID]
		for _, token := range surface.CoversEvidence {
			if covered[loopID] == nil {
				covered[loopID] = map[string]bool{}
			}
			covered[loopID][token] = true
		}
	}
	for _, loop := range m.Loops {
		for _, source := range loop.Evidence {
			if !covered[loop.ID][source.Path] {
				t.Fatalf("claim surface evidence misses %s evidence source %s", loop.ID, source.Path)
			}
		}
	}
}
