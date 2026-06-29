package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestGitHubAnnotationsIncludeClaimVerifierCommands(t *testing.T) {
	var out bytes.Buffer
	writeGitHubAnnotations(&out, verifyResult{
		ClaimSurfaces: []claimSurface{{
			ID: "claim:one",
			VerifierCommands: []string{
				"go test ./tools/loopregistry -run '^(TestClaim)$' -count=1",
			},
		}},
	}, nil)
	got := out.String()
	for _, want := range []string{
		"::notice title=Riido claim verifier::",
		"claim:one => go test ./tools/loopregistry",
		"-run '^(TestClaim)$' -count=1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("annotation missing %q: %s", want, got)
		}
	}
}

func TestGitHubAnnotationsIncludeImpactScope(t *testing.T) {
	var out bytes.Buffer
	writeGitHubAnnotations(&out, verifyResult{}, &impactEvidence{
		Enabled:          true,
		ChangedFileCount: 2,
		ChangedFiles:     []string{"docs/claim.md", "internal/example.go"},
		TargetVerifierPlan: &targetVerifierPlan{
			MatchedPathCount: 1,
			CommandCount:     4,
			VerifierCommands: []string{
				"go test ./tools/a -count=1",
				"go test ./tools/b -count=1",
				"go test ./tools/c -count=1",
				"go test ./tools/d -count=1",
			},
		},
	})
	got := out.String()
	for _, want := range []string{
		"::notice title=Riido impact scope::",
		"2 changed files: docs/claim.md, internal/example.go",
		"target verifiers: 1 matched paths, 4 commands",
		"::notice title=Riido target verifier plan::",
		"commands: go test ./tools/a -count=1",
		"go test ./tools/b -count=1",
		"+2 more in loop-registry-evidence",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("annotation missing %q: %s", want, got)
		}
	}
}

func TestGitHubAnnotationEscapesWorkflowSyntax(t *testing.T) {
	got := githubAnnotationProperty("title: one,two")
	if got != "title%3A one%2Ctwo" {
		t.Fatalf("property escape = %q", got)
	}
	if got := githubAnnotationMessage("a%b\nc"); got != "a%25b%0Ac" {
		t.Fatalf("message escape = %q", got)
	}
}
