package main

import (
	"strings"
	"testing"
)

func TestTargetVerifierPathSummaryUsesMatchedPaths(t *testing.T) {
	got := targetVerifierPathSummaryFor(&targetVerifierPlan{
		Paths: []targetVerifierPath{
			{Path: "tools/loopregistry/z.go"},
			{Path: "docs/30-architecture/loop-registry.riido.json"},
			{Path: "tools/loopregistry/a.go"},
		},
	}, "evidence.json")
	for _, want := range []string{
		"paths: docs/30-architecture/loop-registry.riido.json, tools/loopregistry/a.go",
		"+1 more in evidence.json",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("path summary missing %q: %s", want, got)
		}
	}
}

func TestTargetVerifierPathSummarySkipsMissingPlan(t *testing.T) {
	if got := targetVerifierPathSummaryFor(nil, "evidence.json"); got != "" {
		t.Fatalf("path summary = %q", got)
	}
}
