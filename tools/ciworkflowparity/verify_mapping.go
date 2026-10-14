package main

import "slices"

func verifyNativeMapping(document manifest) bool {
	mapping, claim := document.NativeMapping, document.ParityClaim
	dependencyCommand := "mkdir -p out && go run ./tools/dependencyallowlist -contract dependency_allowlist.riido.json -evidence-out out/dependency-allowlist-evidence.json"
	return mapping.Checkout.Source == "actions/checkout@v7" &&
		mapping.Checkout.NativeKind == "checkout" && !mapping.Checkout.AdapterRequired &&
		mapping.DependencyAllowlist.SourceCommand == dependencyCommand &&
		mapping.DependencyAllowlist.NativeKind == "shell" &&
		mapping.DependencyAllowlist.EvidencePath == "out/dependency-allowlist-evidence.json" &&
		!mapping.DependencyAllowlist.AdapterRequired &&
		mapping.Test.SourceCommand == "go test ./..." && mapping.Test.NativeKind == "shell" &&
		!mapping.Test.AdapterRequired && claim.AllSourceStepsMapped &&
		claim.RequiredAdapterCount == 0 && claim.TestCommandExact && claim.DependencyCommandExact &&
		!claim.SourceWorkflowEdited && !claim.SourceWorkflowExecutedByThisSlice
}

func verifyArtifactMapping(document manifest) bool {
	value, claim := document.NativeMapping.EvidenceArtifact, document.ParityClaim
	want := []string{"out/baseline-ci-parity-evidence.json", "out/dependency-allowlist-evidence.json"}
	return value.Source == "actions/upload-artifact@v7" && value.NativeKind == "artifact" &&
		slices.Equal(value.Paths, want) && value.Redaction == "required" && value.RunWhen == "always" &&
		value.IfNoFilesFound == "error" && value.ContentAddressed && !value.AdapterRequired &&
		claim.DependencyFailurePreservesRedactedArtifactAttempt &&
		claim.LaterOnSuccessTestSkipsAfterPriorFailure
}
