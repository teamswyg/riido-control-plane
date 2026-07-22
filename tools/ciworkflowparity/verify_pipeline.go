package main

import (
	"slices"
	"strings"
)

func verifyPipeline(repoRoot string, document manifest) bool {
	var value pipeline
	if err := readJSON(repoRoot, document.Runner.Pipeline, &value); err != nil {
		return false
	}
	if !validPipelineIdentity(value, document) {
		return false
	}
	steps := value.Steps
	return steps[0].ID == "checkout" && steps[0].Kind == "checkout" &&
		steps[1].ID == "verify-baseline-ci-parity" && steps[1].Kind == "shell" &&
		strings.Contains(steps[1].Command, "go run ./tools/ciworkflowparity") &&
		steps[2].ID == "verify-go-dependency-allowlist" && steps[2].Kind == "shell" &&
		steps[2].Command == document.NativeMapping.DependencyAllowlist.SourceCommand &&
		steps[3].ID == "collect-baseline-policy-evidence" && steps[3].Kind == "artifact" &&
		slices.Equal(steps[3].Paths, document.NativeMapping.EvidenceArtifact.Paths) &&
		steps[3].Redaction == "required" && steps[3].RunWhen == "always" &&
		steps[4].ID == "test-control-plane" && steps[4].Kind == "shell" &&
		steps[4].Command == document.NativeMapping.Test.SourceCommand &&
		slices.Contains(value.Evidence.OwnerPackages, "tools/ciworkflowparity")
}

func validPipelineIdentity(value pipeline, document manifest) bool {
	return value.SchemaVersion == "riido-ci-pipeline.v1" && value.ID == document.Runner.PipelineID &&
		value.Status == "active" && value.Repo == "riido-control-plane" && value.Visibility == "private" &&
		value.Execution.DefaultEngine == "wasm" && value.Execution.NativePolicy == "explicit" &&
		value.Execution.Attestation == "required" && value.Evidence.Artifact != "" &&
		len(value.Steps) == 10 && len(value.Evidence.Cases) == 6 &&
		len(value.Evidence.SourceChecks) == 10 && len(value.SuccessGate) == 12 &&
		completeLoop(value.Evidence.Loop)
}
