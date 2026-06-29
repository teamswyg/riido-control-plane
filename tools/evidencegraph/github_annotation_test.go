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
		ChangedFiles:      []string{"docs/30-architecture/evidence-graph.riido.json", "tools/evidencegraph/run.go"},
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
		"2 changed files (docs/30-architecture/evidence-graph.riido.json, tools/evidencegraph/run.go)",
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

func TestChangedFileAnnotationSuffixLimitsLongLists(t *testing.T) {
	got := changedFileAnnotationSuffix([]string{"a", "b", "c", "d", "e", "f"})
	want := " (a, b, c, d, e (+1 more))"
	if got != want {
		t.Fatalf("suffix = %q, want %q", got, want)
	}
}
