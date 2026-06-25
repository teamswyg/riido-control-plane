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

func TestLoopRegistryWorkflowPublishesRefreshCommandArtifact(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/loop-registry.yml")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "-refresh-plan-in out/loop-registry-evidence.json") ||
		!strings.Contains(text, "-refresh-commands-out out/loop-refresh-commands.json") {
		t.Fatal("loop-registry workflow must derive refresh command evidence from loop evidence")
	}
	if !workflowUploadsStrictArtifact(text, "loop-refresh-commands") {
		t.Fatal("loop-registry workflow must upload strict loop-refresh-commands artifact")
	}
}
