package main

import "slices"

const (
	openQuestionsLoopID           = "open_decision_queue"
	openQuestionsDependencySource = "go run ./tools/dependencyallowlist -contract dependency_allowlist.riido.json"
	openQuestionsDependencyNative = "mkdir -p out && go run ./tools/dependencyallowlist -contract dependency_allowlist.riido.json -evidence-out out/dependency-allowlist-evidence.json"
	openQuestionsSource           = "mkdir -p out && go test ./tools/openquestions -count=1 && go run ./tools/openquestions -check-doc -evidence-out out/open-questions-evidence.json"
	openQuestionsNative           = "umask 077 && export RIIDO_LOOP_IDS=" + openQuestionsLoopID + " && " + openQuestionsSource
	openQuestionsKnowledgeSource  = "go run ./tools/knowledgecoverage -check-doc -evidence-out out/executable-knowledge-coverage.json"
	openQuestionsKnowledgeNative  = "umask 077 && export RIIDO_LOOP_IDS=" + openQuestionsLoopID + " && " + openQuestionsKnowledgeSource
)

func verifyOpenQuestionsWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: open-questions", "RIIDO_LOOP_IDS: open_decision_queue",
		"actions/checkout@v7", "actions/setup-go@v6", "go-version-file: go.mod", "cache: false",
		openQuestionsDependencySource, "go test ./tools/openquestions -count=1",
		"go run ./tools/openquestions", "-evidence-out out/open-questions-evidence.json",
		"go run ./tools/knowledgecoverage", "-evidence-out out/executable-knowledge-coverage.json",
		"actions/upload-artifact@v7", "if: always()", "name: open-questions-evidence",
		"out/open-questions-evidence.json", "out/executable-knowledge-coverage.json",
		"if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "61753756e48290103ab7a20b876fafa747cd000f" &&
		value.Workflow == ".github/workflows/open-questions.yml" &&
		value.WorkflowSHA256 == "a73744d4a386246d66f1c599953b8411a4740ca3616957857ed415ef6500c937" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "open-questions" &&
		value.Job == "open-questions" && value.TrackedWorkflowCount == 56 &&
		containsAll(string(raw), required)
}

func verifyOpenQuestionsMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && !mapping.GoToolchain.Cache &&
		!mapping.GoToolchain.AdapterRequired &&
		verifyCommand(mapping.DependencyAllowlist, openQuestionsDependencySource,
			openQuestionsDependencyNative, "out/dependency-allowlist-evidence.json") &&
		verifyCommand(mapping.OpenQuestions, openQuestionsSource, openQuestionsNative,
			"out/open-questions-evidence.json") &&
		verifyCommand(mapping.ExecutableKnowledge, openQuestionsKnowledgeSource,
			openQuestionsKnowledgeNative, "out/executable-knowledge-coverage.json") &&
		claim.AllSourceStepsMapped && claim.RequiredAdapterCount == 0 &&
		claim.DependencyAllowlistBehaviorExact && claim.OpenQuestionsCommandExact &&
		claim.ExecutableKnowledgeCommandExact && claim.OpenQuestionsLoopEnvironmentExact &&
		claim.SecureEvidencePermissions && !claim.SourceWorkflowEdited &&
		!claim.SourceWorkflowExecutedByThisSlice
}

func verifyOpenQuestionsArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{"out/open-questions-evidence.json", "out/executable-knowledge-coverage.json"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired
}
