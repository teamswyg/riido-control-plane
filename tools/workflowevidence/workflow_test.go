package main

import "testing"

func TestWorkflowEvidenceRequiresSchedule(t *testing.T) {
	if workflowScheduled("on:\n  workflow_dispatch:\n") {
		t.Fatal("workflow without cron schedule must not pass")
	}
	if !workflowScheduled("on:\n  schedule:\n    - cron: \"37 20 * * *\"\n") {
		t.Fatal("workflow with cron schedule should pass")
	}
}

func TestWorkflowEvidenceRequiresStrictUpload(t *testing.T) {
	text := "uses: actions/upload-artifact@v7\nwith:\n  name: workflow-evidence\n"
	if workflowUploadsEvidence(text, "workflow-evidence") {
		t.Fatal("artifact upload without if-no-files-found:error must not pass")
	}
	text += "  if-no-files-found: error\n"
	if !workflowUploadsEvidence(text, "workflow-evidence") {
		t.Fatal("strict workflow-evidence artifact upload should pass")
	}
}
