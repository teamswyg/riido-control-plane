package main

import "slices"

func verifySyntaxHashPipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil || len(value.Steps) != pipelineSteps {
		return false
	}
	steps := value.Steps
	return steps[24].ID == "verify-syntax-hash-golden-locks" && steps[24].Kind == "shell" &&
		steps[24].Command == child.NativeMapping.SyntaxHashGolden.NativeCommand &&
		steps[25].ID == "verify-syntax-hash" && steps[25].Kind == "shell" &&
		steps[25].Command == child.NativeMapping.SyntaxHash.NativeCommand &&
		steps[26].ID == "collect-syntax-hash-evidence" && steps[26].Kind == "artifact" &&
		slices.Equal(steps[26].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[26].Redaction == "required" && steps[26].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/syntaxhash")
}
