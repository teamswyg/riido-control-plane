package main

import "slices"

func verifyEvidenceGraphPipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil ||
		len(value.Steps) != pipelineSteps {
		return false
	}
	steps := value.Steps
	return steps[44].ID == "verify-evidence-graph" && steps[44].Kind == "shell" &&
		steps[44].Command == child.NativeMapping.EvidenceGraph.NativeCommand &&
		steps[45].ID == "collect-evidence-graph-evidence" && steps[45].Kind == "artifact" &&
		slices.Equal(steps[45].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[45].Redaction == "required" && steps[45].RunWhen == "on_success" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/evidencegraph")
}
