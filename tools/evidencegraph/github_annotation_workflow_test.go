package main

import (
	"os"
	"strings"
	"testing"
)

func TestEvidenceGraphWorkflowEmitsGitHubAnnotations(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/evidence-graph.yml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "-github-annotations") {
		t.Fatal("evidence-graph workflow must emit impact annotations")
	}
}
