package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowDispatchPublishesScopeSummary(t *testing.T) {
	text := loopRefreshDispatchWorkflowText(t)
	for _, required := range []string{
		"GITHUB_STEP_SUMMARY",
		"::notice title=Loop refresh dispatch::",
		".claim_ids",
		".evidence_chain_ids",
		".inputs",
		"args+=(\"-f\" \"$pair\")",
		"evidence_chains=",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("workflow must publish dispatch scope: missing %q", required)
		}
	}
}

func loopRefreshDispatchWorkflowText(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRootForTest(t), ".github", "workflows", "loop-refresh-dispatch.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
