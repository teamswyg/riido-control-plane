package main

import "slices"

func verifyHarnessPromotionPipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil ||
		len(value.Steps) != pipelineSteps {
		return false
	}
	steps := value.Steps
	return steps[37].ID == "verify-harness-promotion" && steps[37].Kind == "shell" &&
		steps[37].Command == child.NativeMapping.HarnessPromotion.NativeCommand &&
		steps[38].ID == "collect-harness-promotion-evidence" && steps[38].Kind == "artifact" &&
		slices.Equal(steps[38].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[38].Redaction == "required" && steps[38].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/harnesspromotion")
}
