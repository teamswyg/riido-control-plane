package main

import (
	"os"
	"strings"
	"testing"
)

func TestEvidenceGraphWorkflowIsScheduled(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/evidence-graph.yml")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "schedule:") || !strings.Contains(text, "- cron:") {
		t.Fatal("evidence graph workflow must run on a schedule")
	}
}
