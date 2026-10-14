package main

import "slices"

func verifyOpenQuestionsPipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil ||
		len(value.Steps) != pipelineSteps {
		return false
	}
	steps := value.Steps
	return steps[2].ID == "verify-go-dependency-allowlist" && steps[2].Kind == "shell" &&
		steps[2].Command == child.NativeMapping.DependencyAllowlist.NativeCommand &&
		steps[34].ID == "verify-open-questions" && steps[34].Kind == "shell" &&
		steps[34].Command == child.NativeMapping.OpenQuestions.NativeCommand &&
		steps[35].ID == "verify-open-questions-executable-knowledge" && steps[35].Kind == "shell" &&
		steps[35].Command == child.NativeMapping.ExecutableKnowledge.NativeCommand &&
		steps[36].ID == "collect-open-questions-evidence" && steps[36].Kind == "artifact" &&
		slices.Equal(steps[36].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[36].Redaction == "required" && steps[36].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/openquestions") &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/knowledgecoverage")
}
