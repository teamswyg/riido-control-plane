package main

import "slices"

func verifyWorkflowEvidencePipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil ||
		len(value.Steps) != pipelineSteps {
		return false
	}
	steps := value.Steps
	return steps[32].ID == "verify-workflow-evidence" && steps[32].Kind == "shell" &&
		steps[32].Command == child.NativeMapping.WorkflowEvidence.NativeCommand &&
		steps[33].ID == "collect-workflow-evidence" && steps[33].Kind == "artifact" &&
		slices.Equal(steps[33].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[33].Redaction == "required" && steps[33].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/workflowevidence")
}
