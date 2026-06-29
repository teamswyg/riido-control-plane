package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestGitHubAnnotationsIncludeChainImpactScope(t *testing.T) {
	var out bytes.Buffer
	writeGitHubAnnotations(&out, &impactEvidence{
		Enabled:           true,
		ChangedFileCount:  2,
		ChangedChainCount: 1,
		ChangedChains: []impactChain{{
			ID:                    "chain:one",
			ChangedExecutableRefs: []string{"tools/evidencegraph/run.go"},
			Claims:                []string{"claim:one"},
			NextLoop:              "closed_loop_candidate",
		}},
	})
	got := out.String()
	for _, want := range []string{
		"::notice title=Riido evidence graph impact::",
		"2 changed files, 0 added chains, 1 changed chains, 0 removed chains",
		"::notice title=Riido evidence chain impact::",
		"chain:one executable refs: tools/evidencegraph/run.go claims: claim:one next_loop: closed_loop_candidate",
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
