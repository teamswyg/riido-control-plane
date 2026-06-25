package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestGitHubAnnotationsIncludeClaimImpactScope(t *testing.T) {
	var out bytes.Buffer
	writeGitHubAnnotations(&out, verifyResult{}, &impactEvidence{
		Enabled:          true,
		ChangedFileCount: 1,
		ChangedFiles:     []string{"internal/example.go"},
		AddedClaims:      []impactClaim{{ID: "claim_added"}},
		Claims:           []impactClaim{{ID: "claim_changed"}},
		RemovedClaims:    []impactClaim{{ID: "claim_removed"}},
		BoundSurfaces: []impactBoundSurface{{
			ID:                "claim_bound",
			ChangedBoundFiles: []string{"internal/example.go"},
			ChangedEvidence:   []string{"docs/claim.md"},
		}},
	})
	got := out.String()
	for _, want := range []string{
		"added claims: claim_added",
		"changed claims: claim_changed",
		"removed claims: claim_removed",
		"bound surfaces: claim_bound files: internal/example.go evidence: docs/claim.md",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("annotation missing %q: %s", want, got)
		}
	}
}
