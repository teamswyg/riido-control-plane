package main

import "slices"

func verifyLoopClosureAuditPipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil ||
		len(value.Steps) != pipelineSteps {
		return false
	}
	steps := value.Steps
	return steps[42].ID == "verify-loop-closure-audit" && steps[42].Kind == "shell" &&
		steps[42].Command == child.NativeMapping.LoopClosureAudit.NativeCommand &&
		steps[43].ID == "collect-loop-closure-audit-evidence" && steps[43].Kind == "artifact" &&
		slices.Equal(steps[43].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[43].Redaction == "required" && steps[43].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/loopclosureaudit")
}
