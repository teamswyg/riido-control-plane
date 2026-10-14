package main

import "slices"

func verifyPreCommitPipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil || len(value.Steps) != pipelineSteps {
		return false
	}
	steps := value.Steps
	return steps[19].ID == "verify-pre-commit-baseline" && steps[19].Kind == "shell" &&
		steps[19].Command == child.NativeMapping.PreCommitBaseline.NativeCommand &&
		steps[20].ID == "collect-pre-commit-baseline-evidence" && steps[20].Kind == "artifact" &&
		slices.Equal(steps[20].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[20].Redaction == "required" && steps[20].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/precommitbaseline")
}
