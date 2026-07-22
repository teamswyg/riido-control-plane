package main

import "strings"

func recordPreCommitCases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[4]
	record("pre_commit_baseline_workflow_digest_is_exact", verifyPreCommitWorkflow(repoRoot, child))
	record("pre_commit_baseline_four_source_behaviors_are_native", verifyPreCommitMapping(child))
	record("pre_commit_baseline_artifact_is_redacted", verifyPreCommitArtifact(child))
	record("pre_commit_baseline_pipeline_preserves_order", verifyPreCommitPipeline(repoRoot, document, child))
	record("pre_commit_baseline_authority_remains_zero", verifyAuthorityForChild(child))
	record("pre_commit_baseline_classification_is_bounded", verifyPreCommitClassification(child))
}

func verifyPreCommitClassification(child boundedChild) bool {
	return child.Classification.Code == "pre_commit_baseline_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "pre-commit baseline") &&
		len(child.Classification.DoesNotClaim) == 5
}
