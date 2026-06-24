package main

import "testing"

func TestWorkflowTextUploadsArtifact(t *testing.T) {
	text := "steps:\n- uses: actions/upload-artifact@v7\n  with:\n    name: example-evidence\n"
	if !workflowTextUploadsArtifact(text, "example-evidence") {
		t.Fatal("expected artifact upload match")
	}
}

func TestWorkflowTextUploadsArtifactMatchesNamedStep(t *testing.T) {
	text := "steps:\n- name: Upload evidence\n  uses: actions/upload-artifact@v7\n  with:\n    name: example-evidence\n"
	if !workflowTextUploadsArtifact(text, "example-evidence") {
		t.Fatal("expected named upload step match")
	}
}

func TestWorkflowTextUploadsArtifactStrictRequiresErrorMode(t *testing.T) {
	warn := "steps:\n- uses: actions/upload-artifact@v7\n  with:\n    name: example-evidence\n    if-no-files-found: warn\n"
	if workflowTextUploadsArtifactStrict(warn, "example-evidence") {
		t.Fatal("warn mode must not count as strict evidence upload")
	}
	strict := "steps:\n- uses: actions/upload-artifact@v7\n  with:\n    name: example-evidence\n    if-no-files-found: error\n"
	if !workflowTextUploadsArtifactStrict(strict, "example-evidence") {
		t.Fatal("expected strict artifact upload match")
	}
}

func TestWorkflowTextUploadsArtifactStrictStopsAtNextUpload(t *testing.T) {
	text := "steps:\n- uses: actions/upload-artifact@v7\n  with:\n    name: example-evidence\n- uses: actions/upload-artifact@v7\n  with:\n    name: other-evidence\n    if-no-files-found: error\n"
	if workflowTextUploadsArtifactStrict(text, "example-evidence") {
		t.Fatal("strict mode must not borrow error handling from a later upload")
	}
}

func TestWorkflowTextUploadsArtifactStrictStopsAtNextStep(t *testing.T) {
	text := "steps:\n- uses: actions/upload-artifact@v7\n  with:\n    name: example-evidence\n- run: echo later\n  env:\n    if-no-files-found: error\n"
	if workflowTextUploadsArtifactStrict(text, "example-evidence") {
		t.Fatal("strict mode must not borrow error handling from a later step")
	}
}

func TestWorkflowTextUploadsArtifactPathMatchesSameStep(t *testing.T) {
	text := "steps:\n- uses: actions/upload-artifact@v7\n  with:\n    name: example-evidence\n    path: out/example.json\n"
	if !workflowTextUploadsArtifactPath(text, "example-evidence", "out/example.json") {
		t.Fatal("expected upload path match")
	}
}

func TestWorkflowTextUploadsArtifactPathStopsAtNextStep(t *testing.T) {
	text := "steps:\n- uses: actions/upload-artifact@v7\n  with:\n    name: example-evidence\n- run: echo later\n  env:\n    path: out/example.json\n"
	if workflowTextUploadsArtifactPath(text, "example-evidence", "out/example.json") {
		t.Fatal("upload path must not be borrowed from a later step")
	}
}

func TestWorkflowTextUploadsArtifactPathStrictRequiresSameStep(t *testing.T) {
	text := "steps:\n- uses: actions/upload-artifact@v7\n  with:\n    name: example-evidence\n    path: out/example.json\n- uses: actions/upload-artifact@v7\n  with:\n    name: example-evidence\n    path: out/other.json\n    if-no-files-found: error\n"
	if workflowTextUploadsArtifactPathStrict(text, "example-evidence", "out/example.json") {
		t.Fatal("path and strict missing-file behavior must belong to the same upload step")
	}
}
