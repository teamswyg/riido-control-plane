package main

import "testing"

func TestClaimSurfaceEvidenceIncludesFailureCoverage(t *testing.T) {
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
		for _, token := range surface.CoversFails {
			if covered[loopID] == nil {
				covered[loopID] = map[string]bool{}
			}
			covered[loopID][token] = true
		}
	}
	for _, loop := range m.Loops {
		for _, token := range loop.FailsWhen {
			if !covered[loop.ID][token] {
				t.Fatalf("claim surface evidence misses %s failure token %s", loop.ID, token)
			}
		}
	}
}

func claimLoopIndex(claims []claimBinding) map[string]string {
	out := map[string]string{}
	for _, claim := range claims {
		out[claim.ID] = claim.Loop
	}
	return out
}
