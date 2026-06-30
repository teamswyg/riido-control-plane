package main

import (
	"os"
	"strings"
	"testing"
)

func TestLoopRegistryWorkflowExecutesTargetVerifierScript(t *testing.T) {
	root, err := findRepoRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(repoPath(root, ".github/workflows/loop-registry.yml"))
	if err != nil {
		t.Fatalf("read loop-registry workflow: %v", err)
	}
	text := string(data)
	for _, phrase := range []string{
		"-target-verifier-script-out out/loop-target-verifiers.sh",
		"bash out/loop-target-verifiers.sh",
		"name: loop-target-verifiers",
	} {
		if !strings.Contains(text, phrase) {
			t.Fatalf("loop-registry workflow missing %q", phrase)
		}
	}
}
