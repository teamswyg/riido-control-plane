package main

import "testing"

func TestClaimSurfaceRejectsManifestOnlyClaim(t *testing.T) {
	claim := claimBinding{
		ID:           "manifest_only",
		Statement:    "This claim has no executable code or test path.",
		Files:        []string{"docs/30-architecture/loop-registry.riido.json"},
		Verifiers:    []string{"TestClaimSurfaceRejectsManifestOnlyClaim"},
		GeneratedDoc: []string{"docs/30-architecture/loop-registry.md"},
	}
	if err := verifyClaimSurface(claim); err == nil {
		t.Fatal("expected manifest-only claim to fail")
	}
}
