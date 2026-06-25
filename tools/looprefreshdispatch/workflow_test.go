package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoopRefreshDispatchWorkflowDoesNotEvalCommands(t *testing.T) {
	path := filepath.Join(repoRootForTest(t), ".github", "workflows", "loop-refresh-dispatch.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, forbidden := range []string{"eval ", ".commands[].command", "bash -c"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("workflow must not execute raw command strings: found %q", forbidden)
		}
	}
	if !strings.Contains(text, ".dispatches[].workflow_file") ||
		!strings.Contains(text, `gh workflow run "$workflow" --ref main`) {
		t.Fatalf("workflow must dispatch only structured workflow_file values")
	}
}
