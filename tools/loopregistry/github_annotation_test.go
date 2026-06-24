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
	})
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

func TestGitHubAnnotationEscapesWorkflowSyntax(t *testing.T) {
	got := githubAnnotationProperty("title: one,two")
	if got != "title%3A one%2Ctwo" {
		t.Fatalf("property escape = %q", got)
	}
	if got := githubAnnotationMessage("a%b\nc"); got != "a%25b%0Ac" {
		t.Fatalf("message escape = %q", got)
	}
}
