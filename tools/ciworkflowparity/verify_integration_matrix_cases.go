package main

import "strings"

func recordIntegrationMatrixCases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[12]
	record("integration_matrix_workflow_digest_is_exact", verifyIntegrationMatrixWorkflow(repoRoot, child))
	record("integration_matrix_six_source_behaviors_are_native", verifyIntegrationMatrixMapping(child))
	record("integration_matrix_artifact_is_redacted", verifyIntegrationMatrixArtifact(child))
	record("integration_matrix_pipeline_preserves_order", verifyIntegrationMatrixPipeline(repoRoot, document, child))
	record("integration_matrix_authority_remains_zero", verifyAuthorityForChild(child))
	record("integration_matrix_classification_is_bounded", verifyIntegrationMatrixClassification(child))
}

func verifyIntegrationMatrixClassification(child boundedChild) bool {
	return child.Classification.Code == "integration_matrix_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "integration-matrix") &&
		len(child.Classification.DoesNotClaim) == 5
}
