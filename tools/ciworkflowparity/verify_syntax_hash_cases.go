package main

import "strings"

func recordSyntaxHashCases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[6]
	record("syntax_hash_workflow_digest_is_exact", verifySyntaxHashWorkflow(repoRoot, child))
	record("syntax_hash_five_source_behaviors_are_native", verifySyntaxHashMapping(child))
	record("syntax_hash_artifact_is_redacted", verifySyntaxHashArtifact(child))
	record("syntax_hash_pipeline_preserves_order", verifySyntaxHashPipeline(repoRoot, document, child))
	record("syntax_hash_authority_remains_zero", verifyAuthorityForChild(child))
	record("syntax_hash_classification_is_bounded", verifySyntaxHashClassification(child))
}

func verifySyntaxHashClassification(child boundedChild) bool {
	return child.Classification.Code == "syntax_hash_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "syntax-hash") &&
		len(child.Classification.DoesNotClaim) == 5
}
