package main

import "testing"

const repositoryContract = "config/ci/baseline-go-workflow-parity.riido.json"

func TestVerifyProvesNativeBaselineParityWithoutRetirement(t *testing.T) {
	result, err := verify("../..", repositoryContract)
	if err != nil {
		t.Fatalf("%v: %+v", err, result)
	}
	if result.Decision != "passed" || result.PipelineID != "control-plane-local-self-check" ||
		result.RunnerRevision != "4795cdebfcf4bccaa71ec2344368ad9adf6b1974" ||
		result.RequiredAdapterCount != 0 || !result.LegacyWorkflowPreserved ||
		result.RetirementAuthorized || result.RuntimeEffect != "none" {
		t.Fatalf("unexpected parity evidence: %+v", result)
	}
	expected := []string{
		"control-plane-repository-readme-ci-parity", "control-plane-context-map-ci-parity",
		"control-plane-go-ci-quality-parity", "control-plane-module-decomposition-ci-parity",
		"control-plane-pre-commit-baseline-ci-parity", "control-plane-migration-ledger-ci-parity",
		"control-plane-syntax-hash-ci-parity", "control-plane-config-reference-ci-parity",
		"control-plane-executable-knowledge-ci-parity", "control-plane-workflow-evidence-ci-parity",
		"control-plane-open-questions-ci-parity",
		"control-plane-harness-promotion-ci-parity",
		"control-plane-integration-matrix-ci-parity",
	}
	if len(result.Cases) != 88 || len(result.BoundedChildren) != len(expected) {
		t.Fatalf("unexpected cases: %+v", result.Cases)
	}
	for index, child := range result.BoundedChildren {
		if child.ID != expected[index] || child.RequiredAdapterCount != 0 ||
			!child.LegacyWorkflowPreserved || child.RetirementAuthorized ||
			child.RuntimeEffect != "none" {
			t.Fatalf("unexpected bounded child: %+v", child)
		}
	}
}
