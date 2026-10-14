package main

import "slices"

func verifyIntegrationMatrixPipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil ||
		len(value.Steps) != pipelineSteps {
		return false
	}
	steps := value.Steps
	return steps[2].ID == "verify-go-dependency-allowlist" && steps[2].Kind == "shell" &&
		steps[2].Command == child.NativeMapping.DependencyAllowlist.NativeCommand &&
		steps[39].ID == "verify-integration-matrix" && steps[39].Kind == "shell" &&
		steps[39].Command == child.NativeMapping.IntegrationMatrix.NativeCommand &&
		steps[40].ID == "verify-integration-matrix-executable-knowledge" && steps[40].Kind == "shell" &&
		steps[40].Command == child.NativeMapping.ExecutableKnowledge.NativeCommand &&
		steps[41].ID == "collect-integration-matrix-evidence" && steps[41].Kind == "artifact" &&
		slices.Equal(steps[41].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[41].Redaction == "required" && steps[41].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/integrationmatrix") &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/knowledgecoverage")
}
