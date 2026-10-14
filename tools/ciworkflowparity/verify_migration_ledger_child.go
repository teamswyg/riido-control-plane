package main

const (
	migrationDependencySource = "go run ./tools/dependencyallowlist -contract dependency_allowlist.riido.json"
	migrationDependencyNative = "mkdir -p out && go run ./tools/dependencyallowlist -contract dependency_allowlist.riido.json -evidence-out out/dependency-allowlist-evidence.json"
	migrationLedgerSource     = "mkdir -p out && go test ./tools/migrationledger -count=1 && go run ./tools/migrationledger -check-doc -evidence-out out/migration-ledger-evidence.json"
	migrationLedgerNative     = "umask 077 && " + migrationLedgerSource
	migrationKnowledgeSource  = "go run ./tools/knowledgecoverage -check-doc -evidence-out out/executable-knowledge-coverage.json"
	migrationKnowledgeNative  = "umask 077 && " + migrationKnowledgeSource
)

func verifyMigrationLedgerWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: migration-ledger", "actions/checkout@v7", "actions/setup-go@v6",
		"go-version-file: go.mod", "cache: false", migrationDependencySource,
		"go test ./tools/migrationledger -count=1", "go run ./tools/migrationledger",
		"-check-doc", "-evidence-out out/migration-ledger-evidence.json",
		"go run ./tools/knowledgecoverage", "-evidence-out out/executable-knowledge-coverage.json",
		"actions/upload-artifact@v7", "if: always()", "name: migration-ledger-evidence",
		"out/migration-ledger-evidence.json", "out/executable-knowledge-coverage.json",
		"if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "32d95404c4b9827b7a9d69d6ae67a2279383a298" &&
		value.Workflow == ".github/workflows/migration-ledger.yml" &&
		value.WorkflowSHA256 == "636e8b00f5ed7df631dc5a17d1560b281b0ab6c99be255afb9dafcec971254b1" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "migration-ledger" &&
		value.Job == "migration-ledger" && value.TrackedWorkflowCount == 56 &&
		containsAll(string(raw), required)
}

func verifyMigrationLedgerMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && !mapping.GoToolchain.Cache && !mapping.GoToolchain.AdapterRequired &&
		verifyCommand(mapping.DependencyAllowlist, migrationDependencySource, migrationDependencyNative,
			"out/dependency-allowlist-evidence.json") &&
		verifyCommand(mapping.MigrationLedger, migrationLedgerSource, migrationLedgerNative,
			"out/migration-ledger-evidence.json") &&
		verifyCommand(mapping.ExecutableKnowledge, migrationKnowledgeSource, migrationKnowledgeNative,
			"out/executable-knowledge-coverage.json") &&
		claim.AllSourceStepsMapped && claim.RequiredAdapterCount == 0 &&
		claim.DependencyAllowlistBehaviorExact && claim.MigrationLedgerCommandExact &&
		claim.ExecutableKnowledgeCommandExact && claim.SecureEvidencePermissions &&
		!claim.SourceWorkflowEdited && !claim.SourceWorkflowExecutedByThisSlice
}
