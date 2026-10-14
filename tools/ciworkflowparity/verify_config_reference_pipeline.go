package main

import "slices"

func verifyConfigReferencePipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil || len(value.Steps) != pipelineSteps {
		return false
	}
	steps := value.Steps
	return steps[2].ID == "verify-go-dependency-allowlist" && steps[2].Kind == "shell" &&
		steps[2].Command == child.NativeMapping.DependencyAllowlist.NativeCommand &&
		steps[27].ID == "verify-config-reference" && steps[27].Kind == "shell" &&
		steps[27].Command == child.NativeMapping.ConfigReference.NativeCommand &&
		steps[28].ID == "verify-config-executable-knowledge" && steps[28].Kind == "shell" &&
		steps[28].Command == child.NativeMapping.ExecutableKnowledge.NativeCommand &&
		steps[29].ID == "collect-config-reference-evidence" && steps[29].Kind == "artifact" &&
		slices.Equal(steps[29].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[29].Redaction == "required" && steps[29].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/configreference") &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/knowledgecoverage")
}
