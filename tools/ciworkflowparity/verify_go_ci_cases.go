package main

import "strings"

func recordGoCICases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[2]
	record("go_ci_workflow_digest_is_exact", verifyGoCIWorkflow(repoRoot, child))
	record("go_ci_eight_source_behaviors_are_native", verifyGoCIMapping(child))
	record("go_ci_artifact_is_redacted", verifyGoCIArtifact(child))
	record("go_ci_pipeline_preserves_order", verifyGoCIPipeline(repoRoot, document, child))
	record("go_ci_coverage_outputs_are_exact", verifyGoCICoverageOutputs(child))
	record("go_ci_authority_remains_zero", verifyAuthorityForChild(child))
	record("go_ci_classification_is_bounded", verifyGoCIClassification(child))
}

func verifyGoCIClassification(child boundedChild) bool {
	return child.Classification.Code == "go_ci_quality_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "Go CI quality") &&
		len(child.Classification.DoesNotClaim) == 5
}
