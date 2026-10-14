package main

import "testing"

func TestLoopRegistryEvidenceExposesLoopSurfaces(t *testing.T) {
	m, hashes := loadLoopRegistryForTest(t)
	result, err := verifyAll("../..", m, hashes)
	if err != nil {
		t.Fatal(err)
	}
	got := newEvidence(m, result, nil)
	if len(got.LoopSurfaces) != len(m.Loops) {
		t.Fatalf("loop surfaces = %d, loops = %d", len(got.LoopSurfaces), len(m.Loops))
	}
	surface := loopSurfaceByID(got.LoopSurfaces, "ai_thread_history")
	if surface.ID == "" || len(surface.Observes) == 0 || len(surface.Verifies) == 0 ||
		len(surface.Evidence) == 0 || len(surface.FailsWhen) == 0 {
		t.Fatalf("incomplete loop surface: %+v", surface)
	}
	if surface.ExpiresAfterHours != 24 || surface.RefreshWorkflow == "" {
		t.Fatalf("loop surface expiry/refresh = %+v", surface)
	}
}

func loopSurfaceByID(surfaces []loopSurface, id string) loopSurface {
	for _, surface := range surfaces {
		if surface.ID == id {
			return surface
		}
	}
	return loopSurface{}
}
