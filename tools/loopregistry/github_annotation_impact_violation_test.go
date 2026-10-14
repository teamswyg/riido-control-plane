package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestGitHubAnnotationsIncludeImpactViolations(t *testing.T) {
	var out bytes.Buffer
	writeGitHubAnnotations(&out, verifyResult{}, &impactEvidence{
		Enabled: true,
		Violations: []impactViolation{{
			ClaimID:                   "claim_one",
			Scope:                     "changed_claim",
			Reason:                    "claim changed without evidence",
			RequiredBoundFiles:        []string{"internal/example.go"},
			RequiredEvidence:          []string{"docs/claim.md"},
			RequiredReasoningEvidence: []string{"docs/evidence.riido.json"},
		}},
	})
	got := out.String()
	for _, want := range []string{
		"::error title=Riido impact violation::",
		"changed_claim:claim_one claim changed without evidence",
		"required files: internal/example.go",
		"required evidence: docs/claim.md",
		"required reasoning: docs/evidence.riido.json",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("annotation missing %q: %s", want, got)
		}
	}
}
