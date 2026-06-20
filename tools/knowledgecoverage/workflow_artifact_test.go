package main

import "testing"

func TestWorkflowTextUploadsArtifact(t *testing.T) {
	text := "steps:\n- uses: actions/upload-artifact@v4\n  with:\n    name: example-evidence\n"
	if !workflowTextUploadsArtifact(text, "example-evidence") {
		t.Fatal("expected artifact upload match")
	}
}

func TestWorkflowTextUploadsArtifactMatchesNamedStep(t *testing.T) {
	text := "steps:\n- name: Upload evidence\n  uses: actions/upload-artifact@v4\n  with:\n    name: example-evidence\n"
	if !workflowTextUploadsArtifact(text, "example-evidence") {
		t.Fatal("expected named upload step match")
	}
}

func TestWorkflowTextUploadsArtifactStrictRequiresErrorMode(t *testing.T) {
	warn := "steps:\n- uses: actions/upload-artifact@v4\n  with:\n    name: example-evidence\n    if-no-files-found: warn\n"
	if workflowTextUploadsArtifactStrict(warn, "example-evidence") {
		t.Fatal("warn mode must not count as strict evidence upload")
	}
	strict := "steps:\n- uses: actions/upload-artifact@v4\n  with:\n    name: example-evidence\n    if-no-files-found: error\n"
	if !workflowTextUploadsArtifactStrict(strict, "example-evidence") {
		t.Fatal("expected strict artifact upload match")
	}
}

func TestWorkflowTextUploadsArtifactStrictStopsAtNextUpload(t *testing.T) {
	text := "steps:\n- uses: actions/upload-artifact@v4\n  with:\n    name: example-evidence\n- uses: actions/upload-artifact@v4\n  with:\n    name: other-evidence\n    if-no-files-found: error\n"
	if workflowTextUploadsArtifactStrict(text, "example-evidence") {
		t.Fatal("strict mode must not borrow error handling from a later upload")
	}
}

func TestWorkflowTextUploadsArtifactStrictStopsAtNextStep(t *testing.T) {
	text := "steps:\n- uses: actions/upload-artifact@v4\n  with:\n    name: example-evidence\n- run: echo later\n  env:\n    if-no-files-found: error\n"
	if workflowTextUploadsArtifactStrict(text, "example-evidence") {
		t.Fatal("strict mode must not borrow error handling from a later step")
	}
}
