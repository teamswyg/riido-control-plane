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
	if len(result.Cases) != 52 || len(result.BoundedChildren) != 7 ||
		result.BoundedChildren[0].ID != "control-plane-repository-readme-ci-parity" ||
		result.BoundedChildren[0].RequiredAdapterCount != 0 ||
		!result.BoundedChildren[0].LegacyWorkflowPreserved ||
		result.BoundedChildren[0].RetirementAuthorized ||
		result.BoundedChildren[0].RuntimeEffect != "none" ||
		result.BoundedChildren[1].ID != "control-plane-context-map-ci-parity" ||
		result.BoundedChildren[1].RequiredAdapterCount != 0 ||
		!result.BoundedChildren[1].LegacyWorkflowPreserved ||
		result.BoundedChildren[1].RetirementAuthorized ||
		result.BoundedChildren[1].RuntimeEffect != "none" ||
		result.BoundedChildren[2].ID != "control-plane-go-ci-quality-parity" ||
		result.BoundedChildren[2].RequiredAdapterCount != 0 ||
		!result.BoundedChildren[2].LegacyWorkflowPreserved ||
		result.BoundedChildren[2].RetirementAuthorized ||
		result.BoundedChildren[2].RuntimeEffect != "none" ||
		result.BoundedChildren[3].ID != "control-plane-module-decomposition-ci-parity" ||
		result.BoundedChildren[3].RequiredAdapterCount != 0 ||
		!result.BoundedChildren[3].LegacyWorkflowPreserved ||
		result.BoundedChildren[3].RetirementAuthorized ||
		result.BoundedChildren[3].RuntimeEffect != "none" ||
		result.BoundedChildren[4].ID != "control-plane-pre-commit-baseline-ci-parity" ||
		result.BoundedChildren[4].RequiredAdapterCount != 0 ||
		!result.BoundedChildren[4].LegacyWorkflowPreserved ||
		result.BoundedChildren[4].RetirementAuthorized ||
		result.BoundedChildren[4].RuntimeEffect != "none" ||
		result.BoundedChildren[5].ID != "control-plane-migration-ledger-ci-parity" ||
		result.BoundedChildren[5].RequiredAdapterCount != 0 ||
		!result.BoundedChildren[5].LegacyWorkflowPreserved ||
		result.BoundedChildren[5].RetirementAuthorized ||
		result.BoundedChildren[5].RuntimeEffect != "none" ||
		result.BoundedChildren[6].ID != "control-plane-syntax-hash-ci-parity" ||
		result.BoundedChildren[6].RequiredAdapterCount != 0 ||
		!result.BoundedChildren[6].LegacyWorkflowPreserved ||
		result.BoundedChildren[6].RetirementAuthorized ||
		result.BoundedChildren[6].RuntimeEffect != "none" {
		t.Fatalf("unexpected cases: %+v", result.Cases)
	}
}
