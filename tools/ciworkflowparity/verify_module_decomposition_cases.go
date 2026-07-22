package main

import "strings"

func recordModuleDecompositionCases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[3]
	record("module_decomposition_workflow_digest_is_exact", verifyModuleDecompositionWorkflow(repoRoot, child))
	record("module_decomposition_six_source_behaviors_are_native", verifyModuleDecompositionMapping(child))
	record("module_decomposition_artifact_is_redacted", verifyModuleDecompositionArtifact(child))
	record("module_decomposition_pipeline_preserves_order", verifyModuleDecompositionPipeline(repoRoot, document, child))
	record("module_decomposition_authority_remains_zero", verifyAuthorityForChild(child))
	record("module_decomposition_classification_is_bounded", verifyModuleDecompositionClassification(child))
}

func verifyModuleDecompositionClassification(child boundedChild) bool {
	return child.Classification.Code == "module_decomposition_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "module-decomposition") &&
		len(child.Classification.DoesNotClaim) == 5
}
