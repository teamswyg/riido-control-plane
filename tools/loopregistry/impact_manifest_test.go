package main

import "testing"

func TestClaimImpactRejectsManifestOnlyChangeForMeaningChange(t *testing.T) {
	base := []claimBinding{manifestOnlyImpactClaim("old")}
	current := []claimBinding{manifestOnlyImpactClaim("new")}
	_, err := verifyClaimImpact("origin/main", base, current, map[string]bool{
		"docs/domain.riido.json": true,
		defaultManifest:          true,
	})
	if err == nil {
		t.Fatal("expected manifest-only claim meaning change to fail")
	}
}

func manifestOnlyImpactClaim(statement string) claimBinding {
	return claimBinding{
		ID:        "claim",
		Statement: statement,
		Loop:      "loop",
		Files: []string{
			"docs/domain.riido.json",
			"internal/example.go",
			"internal/example_test.go",
		},
	}
}
