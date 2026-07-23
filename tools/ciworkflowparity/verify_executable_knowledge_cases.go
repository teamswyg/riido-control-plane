package main

import "strings"

func recordExecutableKnowledgeCases(repoRoot string, document manifest, record func(string, bool)) {
	child := document.BoundedChildren[8]
	record("executable_knowledge_workflow_digest_is_exact", verifyExecutableKnowledgeWorkflow(repoRoot, child))
	record("executable_knowledge_four_source_behaviors_are_native", verifyExecutableKnowledgeMapping(child))
	record("executable_knowledge_artifact_is_redacted", verifyExecutableKnowledgeArtifact(child))
	record("executable_knowledge_pipeline_preserves_order", verifyExecutableKnowledgePipeline(repoRoot, document, child))
	record("executable_knowledge_authority_remains_zero", verifyAuthorityForChild(child))
	record("executable_knowledge_classification_is_bounded", verifyExecutableKnowledgeClassification(child))
}

func verifyExecutableKnowledgeClassification(child boundedChild) bool {
	return child.Classification.Code == "executable_knowledge_native_parity_source_complete" &&
		strings.Contains(child.Classification.Meaning, "executable-knowledge") &&
		len(child.Classification.DoesNotClaim) == 5
}
