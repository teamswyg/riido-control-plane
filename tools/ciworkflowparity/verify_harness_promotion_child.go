package main

import "slices"

const (
	harnessPromotionLoopID = "closed_loop_candidate"
	harnessPromotionSource = "mkdir -p out && go test ./tools/harnesspromotion -count=1 && go run ./tools/harnesspromotion -check-doc -evidence-out out/harness-promotion-evidence.json"
	harnessPromotionNative = "umask 077 && export RIIDO_LOOP_IDS=" + harnessPromotionLoopID + " && " + harnessPromotionSource
)

func verifyHarnessPromotionWorkflow(repoRoot string, child boundedChild) bool {
	value := child.Baseline
	raw, err := readRootFile(repoRoot, value.Workflow)
	required := []string{
		"name: harness-promotion", "workflow_dispatch:", "schedule:", "cron: \"41 20 * * *\"",
		"RIIDO_LOOP_IDS: closed_loop_candidate", "actions/checkout@v7", "actions/setup-go@v6",
		"go-version-file: go.mod", "cache: false", "mkdir -p out",
		"go test ./tools/harnesspromotion -count=1", "go run ./tools/harnesspromotion", "-check-doc",
		"-evidence-out out/harness-promotion-evidence.json", "actions/upload-artifact@v7",
		"if: always()", "name: harness-promotion-evidence",
		"path: out/harness-promotion-evidence.json", "if-no-files-found: error",
	}
	return err == nil && value.SourceRevision == "2d5e59082d9304ecb6393f0d004c75a54626ffe3" &&
		value.Workflow == ".github/workflows/harness-promotion.yml" &&
		value.WorkflowSHA256 == "ae94e1d9332d84e5d1f71ee0b7daa22e110d4a32342d8e7b3e874f3c36b0f762" &&
		digest(raw) == value.WorkflowSHA256 && value.WorkflowName == "harness-promotion" &&
		value.Job == "harness-promotion" && value.TrackedWorkflowCount == 56 &&
		containsAll(string(raw), required)
}

func verifyHarnessPromotionMapping(child boundedChild) bool {
	mapping, claim := child.NativeMapping, child.ParityClaim
	return mapping.Checkout.Source == "actions/checkout@v7" && mapping.Checkout.NativeKind == "checkout" &&
		!mapping.Checkout.AdapterRequired && mapping.GoToolchain.Source == "actions/setup-go@v6" &&
		mapping.GoToolchain.NativeKind == "worker_toolchain" && mapping.GoToolchain.VersionSource == "go.mod" &&
		mapping.GoToolchain.Version == "1.26.5" && !mapping.GoToolchain.Cache &&
		!mapping.GoToolchain.AdapterRequired &&
		verifyCommand(mapping.HarnessPromotion, harnessPromotionSource, harnessPromotionNative,
			"out/harness-promotion-evidence.json") &&
		claim.AllSourceStepsMapped && claim.RequiredAdapterCount == 0 &&
		claim.HarnessPromotionCommandExact && claim.HarnessPromotionLoopEnvironmentExact &&
		claim.SecureEvidencePermissions && !claim.SourceWorkflowEdited &&
		!claim.SourceWorkflowExecutedByThisSlice
}

func verifyHarnessPromotionArtifact(child boundedChild) bool {
	value := child.NativeMapping.EvidenceArtifact
	want := []string{"out/harness-promotion-evidence.json"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired
}
