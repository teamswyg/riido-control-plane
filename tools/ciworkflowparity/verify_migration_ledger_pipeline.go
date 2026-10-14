package main

import "slices"

func verifyMigrationLedgerArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{"out/migration-ledger-evidence.json", "out/executable-knowledge-coverage.json"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired
}

func verifyMigrationLedgerPipeline(repoRoot string, document manifest, child boundedChild) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil || len(value.Steps) != pipelineSteps {
		return false
	}
	steps := value.Steps
	return steps[2].ID == "verify-go-dependency-allowlist" && steps[2].Kind == "shell" &&
		steps[2].Command == child.NativeMapping.DependencyAllowlist.NativeCommand &&
		steps[21].ID == "verify-migration-ledger" && steps[21].Kind == "shell" &&
		steps[21].Command == child.NativeMapping.MigrationLedger.NativeCommand &&
		steps[22].ID == "verify-migration-executable-knowledge" && steps[22].Kind == "shell" &&
		steps[22].Command == child.NativeMapping.ExecutableKnowledge.NativeCommand &&
		steps[23].ID == "collect-migration-ledger-evidence" && steps[23].Kind == "artifact" &&
		slices.Equal(steps[23].Paths, child.NativeMapping.EvidenceArtifact.Paths) &&
		steps[23].Redaction == "required" && steps[23].RunWhen == "always" &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/migrationledger") &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/knowledgecoverage")
}
