package main

import "testing"

func TestWorkflowEvidenceRejectsDeprecatedActions(t *testing.T) {
	record := workflowRecord{
		Path:                  ".github/workflows/example.yml",
		HasExecutable:         true,
		DeprecatedActionRefs:  []string{"actions/checkout@v4"},
		DeprecatedActionCount: 1,
	}
	got := classify(record, nil, nil)
	if got.Status != "deprecated_action" {
		t.Fatalf("expected deprecated_action, got %q", got.Status)
	}
	var result auditResult
	addRecord(&result, got)
	if err := verifyResult(result); err == nil {
		t.Fatal("expected deprecated action verification failure")
	}
}
