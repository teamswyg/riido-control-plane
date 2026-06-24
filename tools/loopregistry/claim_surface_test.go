package main

import "testing"

func TestClaimSurfaceEvidenceIncludesCodeTestDocBindings(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	m, err := loadManifest(repoPath(root, defaultManifest))
	if err != nil {
		t.Fatal(err)
	}
	hashes, err := claimHashes(root, m)
	if err != nil {
		t.Fatal(err)
	}
	result, err := verifyAll(root, m, hashes)
	if err != nil {
		t.Fatalf("verifyAll: %v", err)
	}
	if len(result.ClaimSurfaces) != len(m.Claims) {
		t.Fatalf("claim surfaces = %d, want %d", len(result.ClaimSurfaces), len(m.Claims))
	}
	for _, surface := range result.ClaimSurfaces {
		if len(surface.CodePaths)+len(surface.ManifestPaths) == 0 ||
			len(surface.TestPaths)+len(surface.Verifiers) == 0 ||
			len(surface.GeneratedDocs) == 0 ||
			len(surface.VerifierCommands) == 0 {
			t.Fatalf("incomplete claim surface: %+v", surface)
		}
	}
}

func TestClaimSurfaceRejectsNarrativeOnlyClaim(t *testing.T) {
	claim := claimBinding{
		ID:           "narrative_only",
		Statement:    "This claim has no executable surface.",
		GeneratedDoc: []string{"docs/30-architecture/loop-registry.md"},
	}
	if err := verifyClaimSurface(claim); err == nil {
		t.Fatal("expected narrative-only claim to fail")
	}
}

func TestClaimSurfaceEvidenceIncludesTargetVerifierCommands(t *testing.T) {
	claim := claimBinding{
		ID:        "claim",
		Verifiers: []string{"TestGeneratedDocMatchesManifest"},
		Files: []string{
			"tools/aiagentclientapi/main_test.go",
		},
	}
	tests := map[string][]string{
		"TestGeneratedDocMatchesManifest": {
			"./tools/aiagentclientapi",
			"./tools/configreference",
		},
	}
	surface := claimSurfaceFor(claim, tests)
	want := "go test ./tools/aiagentclientapi -run '^(TestGeneratedDocMatchesManifest)$' -count=1"
	if len(surface.VerifierCommands) != 1 || surface.VerifierCommands[0] != want {
		t.Fatalf("commands = %#v, want %q", surface.VerifierCommands, want)
	}
}
