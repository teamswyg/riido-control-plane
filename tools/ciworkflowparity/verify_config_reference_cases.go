package main

import "strings"

func recordConfigReferenceCases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[7]
	record("config_reference_workflow_digest_is_exact", verifyConfigReferenceWorkflow(repoRoot, child))
	record("config_reference_six_source_behaviors_are_native", verifyConfigReferenceMapping(child))
	record("config_reference_artifact_is_redacted", verifyConfigReferenceArtifact(child))
	record("config_reference_pipeline_preserves_order", verifyConfigReferencePipeline(repoRoot, document, child))
	record("config_reference_authority_remains_zero", verifyAuthorityForChild(child))
	record("config_reference_classification_is_bounded", verifyConfigReferenceClassification(child))
}

func verifyConfigReferenceClassification(child boundedChild) bool {
	return child.Classification.Code == "config_reference_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "config-reference") &&
		len(child.Classification.DoesNotClaim) == 5
}
