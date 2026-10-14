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
		AddedClaims: []impactClaim{{
			ID:                       "claim_added",
			ChangedEvidence:          []string{"docs/added.md"},
			ChangedReasoningEvidence: []string{"docs/reasoning.riido.json"},
		}},
		Claims: []impactClaim{{
			ID:                       "claim_changed",
			ChangedEvidence:          []string{"docs/changed.md"},
			ChangedReasoningEvidence: []string{"docs/reasoning.riido.json"},
		}},
		RemovedClaims: []impactClaim{{
			ID:                       "claim_removed",
			ChangedEvidence:          []string{"docs/removed.md"},
			ChangedReasoningEvidence: []string{"docs/reasoning.riido.json"},
		}},
		BoundSurfaces: []impactBoundSurface{{
			ID:                       "claim_bound",
			ChangedBoundFiles:        []string{"internal/example.go"},
			ChangedEvidence:          []string{"docs/claim.md"},
			ChangedReasoningEvidence: []string{"docs/reasoning.riido.json"},
		}},
	})
	got := out.String()
	for _, want := range []string{
		"added claims: claim_added evidence: docs/added.md reasoning: docs/reasoning.riido.json",
		"changed claims: claim_changed evidence: docs/changed.md reasoning: docs/reasoning.riido.json",
		"removed claims: claim_removed evidence: docs/removed.md reasoning: docs/reasoning.riido.json",
		"bound surfaces: claim_bound files: internal/example.go evidence: docs/claim.md reasoning: docs/reasoning.riido.json",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("annotation missing %q: %s", want, got)
		}
	}
}
