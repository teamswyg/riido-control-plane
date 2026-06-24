package main

import "testing"

func TestWorkflowTextEvidenceOutPathsMatchesBlockCommand(t *testing.T) {
	text := "steps:\n- run: |\n    go run ./tools/example \\\n      -check-doc \\\n      -evidence-out out/example.json\n"
	paths := workflowTextEvidenceOutPaths(text, "./tools/example")
	if len(paths) != 1 || paths[0] != "out/example.json" {
		t.Fatalf("paths = %#v", paths)
	}
}

func TestWorkflowTextUploadsArtifactPathMatchesBlockPath(t *testing.T) {
	text := "steps:\n- uses: actions/upload-artifact@v7\n  with:\n    name: example-evidence\n    path: |\n      out/example.json\n      out/other.json\n"
	if !workflowTextUploadsArtifactPath(text, "example-evidence", "out/example.json") {
		t.Fatal("expected block upload path match")
	}
}
