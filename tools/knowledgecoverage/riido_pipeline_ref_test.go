package main

import "testing"

func TestRiidoPipelineReferenceRequiresStrictSourceOwner(t *testing.T) {
	root := t.TempDir()
	source := sourceSSOT{
		Path: "config/policy.riido.json", EvidenceTool: "tools/policy",
		Workflow: "pipelines/check.riido.json", EvidenceArtifact: "policy-evidence",
	}
	mustWrite(t, root, source.Path, `{"id":"policy","pipelines":["pipelines/check.riido.json"],"assertions":["a"],"loop":{"observation":"o","hypothesis":"h","execute":"x","evaluate":"e","retrospective":"r"}}`)
	mustWrite(t, root, source.Workflow, riidoPipelineFixture("required", "required", "always"))
	m := manifest{SourceManifests: []sourceSSOT{source}, RiidoPipelines: []riidoPipelineRef{{source.Workflow, source.Path, "pipelines"}}}
	if problems := validateRiidoPipelines(root, m); len(problems) != 0 {
		t.Fatalf("strict owned riido-ci route rejected: %v", problems)
	}
}
