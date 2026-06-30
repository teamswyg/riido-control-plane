package main

import (
	"strings"
	"testing"
)

func TestTargetVerifierRunnableSummaryBoundsOutput(t *testing.T) {
	got := targetVerifierRunnableSummaryFor(&targetVerifierPlan{
		RunnableCommands: []string{
			"go test ./tools/a -count=1",
			"go test ./tools/b -count=1",
			"go test ./tools/c -count=1",
		},
	}, "evidence.json")
	for _, want := range []string{
		"runnable: go test ./tools/a -count=1",
		"go test ./tools/b -count=1",
		"+1 more in evidence.json",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("summary missing %q: %s", want, got)
		}
	}
}
