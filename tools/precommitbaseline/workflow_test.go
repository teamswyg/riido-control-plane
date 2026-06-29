package main

import "testing"

func TestPreCommitBaselineWorkflowRequiresSchedule(t *testing.T) {
	if workflowScheduled("on:\n  workflow_dispatch:\n") {
		t.Fatal("workflow without cron schedule must not pass")
	}
	if !workflowScheduled("on:\n  schedule:\n    - cron: \"43 20 * * *\"\n") {
		t.Fatal("workflow with cron schedule should pass")
	}
}

func TestPreCommitBaselineWorkflowRequiresStrictEvidenceUpload(t *testing.T) {
	text := "uses: actions/upload-artifact@v7\nname: pre-commit-baseline-evidence\nif-no-files-found: error"
	if !workflowUploadsEvidence(text, "pre-commit-baseline-evidence") {
		t.Fatal("strict upload should pass")
	}
	if workflowUploadsEvidence(text, "missing-evidence") {
		t.Fatal("missing artifact name should fail")
	}
	if workflowUploadsEvidence("uses: actions/upload-artifact@v7\nname: pre-commit-baseline-evidence", "pre-commit-baseline-evidence") {
		t.Fatal("upload without if-no-files-found error should fail")
	}
}
