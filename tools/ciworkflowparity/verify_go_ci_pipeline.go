package main

import "slices"

func verifyGoCIArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{"out/go-ci-baseline-evidence.json", "out/go-test-cover.out", "out/go-test-coverage-summary.txt"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired
}

func verifyGoCICoverageOutputs(child boundedChild) bool {
	value := child.NativeMapping.Coverage
	return value.EvidencePath == "out/go-test-cover.out" &&
		value.SourceCommand == goCICoverageSourceCommand && value.NativeCommand == goCICoverageNativeCommand &&
		slices.Contains(child.NativeMapping.EvidenceArtifact.Paths, "out/go-test-coverage-summary.txt")
}

func verifyGoCIPipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil || len(value.Steps) != 16 {
		return false
	}
	steps := value.Steps
	return steps[10].ID == "download-go-ci-modules" && steps[10].Kind == "shell" &&
		steps[10].Command == child.NativeMapping.ModuleDownload.NativeCommand &&
		steps[11].ID == "verify-go-ci-baseline" && steps[11].Kind == "shell" &&
		steps[11].Command == child.NativeMapping.GoCIBaseline.NativeCommand &&
		steps[12].ID == "install-go-ci-lint" && steps[12].Kind == "shell" &&
		steps[12].Command == child.NativeMapping.LintInstall.NativeCommand &&
		steps[13].ID == "lint-go-ci" && steps[13].Kind == "shell" &&
		steps[13].Command == child.NativeMapping.Lint.NativeCommand &&
		steps[14].ID == "test-go-ci-with-coverage" && steps[14].Kind == "shell" &&
		steps[14].Command == child.NativeMapping.Coverage.NativeCommand &&
		steps[15].ID == "collect-go-ci-quality-evidence" && steps[15].Kind == "artifact" &&
		slices.Equal(steps[15].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[15].Redaction == "required" && steps[15].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/gocibaseline")
}
