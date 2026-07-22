package main

import "slices"

func verifyReadmeArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{"out/repository-readme-evidence.json", "out/executable-knowledge-coverage.json"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired
}

func verifyReadmePipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil || len(value.Steps) != 16 {
		return false
	}
	steps := value.Steps
	return steps[5].ID == "verify-repository-readme" && steps[5].Kind == "shell" &&
		steps[5].Command == child.NativeMapping.RepositoryReadme.NativeCommand &&
		steps[6].ID == "verify-executable-knowledge" && steps[6].Kind == "shell" &&
		steps[6].Command == child.NativeMapping.ExecutableKnowledge.NativeCommand &&
		steps[7].ID == "collect-repository-readme-evidence" && steps[7].Kind == "artifact" &&
		slices.Equal(steps[7].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[7].Redaction == "required" && steps[7].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/repositoryreadme") &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/knowledgecoverage")
}
