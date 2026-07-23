package main

import "slices"

func verifyExecutableKnowledgePipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil ||
		len(value.Steps) != pipelineSteps {
		return false
	}
	steps := value.Steps
	return steps[30].ID == "verify-dedicated-executable-knowledge" && steps[30].Kind == "shell" &&
		steps[30].Command == child.NativeMapping.ExecutableKnowledge.NativeCommand &&
		steps[31].ID == "collect-dedicated-executable-knowledge-evidence" &&
		steps[31].Kind == "artifact" &&
		slices.Equal(steps[31].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[31].Redaction == "required" && steps[31].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/knowledgecoverage")
}
