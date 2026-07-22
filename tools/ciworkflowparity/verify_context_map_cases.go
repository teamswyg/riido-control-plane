package main

import "strings"

func recordContextMapCases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[1]
	record("context_map_workflow_digest_is_exact", verifyContextMapWorkflow(repoRoot, child))
	record("context_map_steps_are_native", verifyContextMapMapping(child))
	record("context_map_artifact_is_redacted", verifyContextMapArtifact(child))
	record("context_map_pipeline_preserves_order", verifyContextMapPipeline(repoRoot, document, child))
	record("context_map_authority_remains_zero", verifyAuthorityForChild(child))
	record("context_map_classification_is_bounded", verifyContextMapClassification(child))
}

func verifyContextMapClassification(child boundedChild) bool {
	return child.Classification.Code == "context_map_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "Context Map") &&
		len(child.Classification.DoesNotClaim) == 5
}
