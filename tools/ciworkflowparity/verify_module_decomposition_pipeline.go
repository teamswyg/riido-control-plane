package main

import "slices"

func verifyModuleDecompositionArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{"out/module-decomposition-evidence.json", "out/executable-knowledge-coverage.json"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired
}

func verifyModuleDecompositionPipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil || len(value.Steps) != pipelineSteps {
		return false
	}
	steps := value.Steps
	return steps[2].ID == "verify-go-dependency-allowlist" && steps[2].Kind == "shell" &&
		steps[2].Command == child.NativeMapping.DependencyAllowlist.NativeCommand &&
		steps[16].ID == "verify-module-decomposition" && steps[16].Kind == "shell" &&
		steps[16].Command == child.NativeMapping.ModuleDecomposition.NativeCommand &&
		steps[17].ID == "verify-module-executable-knowledge" && steps[17].Kind == "shell" &&
		steps[17].Command == child.NativeMapping.ExecutableKnowledge.NativeCommand &&
		steps[18].ID == "collect-module-decomposition-evidence" && steps[18].Kind == "artifact" &&
		slices.Equal(steps[18].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[18].Redaction == "required" && steps[18].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/moduledecomposition") &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/knowledgecoverage")
}
