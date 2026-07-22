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
	if len(result.Cases) != 21 || len(result.BoundedChildren) != 2 ||
		result.BoundedChildren[0].ID != "control-plane-repository-readme-ci-parity" ||
		result.BoundedChildren[0].RequiredAdapterCount != 0 ||
		!result.BoundedChildren[0].LegacyWorkflowPreserved ||
		result.BoundedChildren[0].RetirementAuthorized ||
		result.BoundedChildren[0].RuntimeEffect != "none" ||
		result.BoundedChildren[1].ID != "control-plane-context-map-ci-parity" ||
		result.BoundedChildren[1].RequiredAdapterCount != 0 ||
		!result.BoundedChildren[1].LegacyWorkflowPreserved ||
		result.BoundedChildren[1].RetirementAuthorized ||
		result.BoundedChildren[1].RuntimeEffect != "none" {
		t.Fatalf("unexpected cases: %+v", result.Cases)
	}
}
