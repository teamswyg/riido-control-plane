package main

import (
	"os"
	"strings"
	"testing"
)

func TestLoopRegistryWorkflowEmitsGitHubAnnotations(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/loop-registry.yml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "-github-annotations") {
		t.Fatal("loop-registry workflow must emit claim verifier annotations")
	}
}
