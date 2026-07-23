package main

import "strings"

func recordHarnessPromotionCases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[11]
	record("harness_promotion_workflow_digest_is_exact", verifyHarnessPromotionWorkflow(repoRoot, child))
	record("harness_promotion_five_source_behaviors_are_native", verifyHarnessPromotionMapping(child))
	record("harness_promotion_artifact_is_redacted", verifyHarnessPromotionArtifact(child))
	record("harness_promotion_pipeline_preserves_order", verifyHarnessPromotionPipeline(repoRoot, document, child))
	record("harness_promotion_authority_remains_zero", verifyAuthorityForChild(child))
	record("harness_promotion_classification_is_bounded", verifyHarnessPromotionClassification(child))
}

func verifyHarnessPromotionClassification(child boundedChild) bool {
	return child.Classification.Code == "harness_promotion_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "harness-promotion") &&
		len(child.Classification.DoesNotClaim) == 5
}
